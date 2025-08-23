import sys
import os
import json
import tempfile
import re
from pathlib import Path

# This script is part of a two-pass build system.
# 1. LaTeX runs first, creating a .scores.aux file with a list of required scores.
# 2. This script reads that .scores.aux file and renders the scores.
# 3. LaTeX runs again to include the newly created score files.

# Import rendering libraries
import verovio
import cairosvg
from PyPDF2 import PdfMerger
import subprocess

def parse_latex_dimension(dim_str: str) -> float:
    """Parses a LaTeX dimension string (e.g., '430.0pt') and returns the float value in points."""
    match = re.match(r"([0-9.]+)", dim_str)
    if match:
        return float(match.group(1))
    return 0.0 # Default or error value

def pt_to_mm(pt_value: float) -> float:
    """Converts a value from points (pt) to millimeters (mm)."""
    return pt_value * 0.3527

def render_score(input_file: Path, output_pdf: Path, verovio_options: dict):
    """Renders a single file to a PDF using a specific set of Verovio options."""

    # No file exit
    if not input_file.is_file():
        print(f"  -> Error: file not found at {input_file}")
        return 0

    # Check if input_file is a .mscz file and convert to .mxl using MuseScore
    temp_mxl = None
    if input_file.suffix == '.mscz':
        try:
            temp_mxl = tempfile.NamedTemporaryFile(delete=False, suffix=".mxl")
            temp_mxl.close()
            result = subprocess.run(["mscore", str(input_file), "-o", temp_mxl.name], capture_output=True, text=True)
            if result.returncode != 0 or not Path(temp_mxl.name).is_file():
                print(f"  -> Error: Failed to convert {input_file} to mxl. MuseScore output:\n{result.stderr}")
                return 0
            input_file = Path(temp_mxl.name)
        except Exception as e:
            print(f"  -> Error: Exception during conversion of {input_file} to mxl: {e}")
            return 0

    # Loading the options and file
    try:
        tk = verovio.toolkit()
        tk.setOptions(verovio_options)
        tk.loadFile(str(input_file))
        page_count = tk.getPageCount()
    except Exception as e: # verovio error exit
        print(f"  -> Error processing {input_file.name}: {e}")
        return 0
    finally:
        if temp_mxl is not None:
            try:
                os.unlink(temp_mxl.name)
            except Exception:
                pass

    # file but no output exit
    if page_count == 0:
        return 0

    # render to svg -> convert to pdf -> append to pages
    with tempfile.TemporaryDirectory() as tmpdir:
        pages = []
        for i in range(1, page_count + 1):
            try:
                svg = tk.renderToSVG(i)
                pdf_page = Path(tmpdir) / f"p{i}.pdf"
                cairosvg.svg2pdf(bytestring=svg.encode('utf-8'), write_to=str(pdf_page))
                pages.append(str(pdf_page))
            except Exception as e:
                print(f"  -> Error converting page {i} of {input_file.name} to PDF: {e}")
                return 0
        
        # Merge Pages to single pdf write to output_pdf path
        merger = PdfMerger()
        for p in pages:
            merger.append(p)
        merger.write(str(output_pdf))
        merger.close()
    
    return page_count

def main():
    # argument handling
    if len(sys.argv) != 2:
        sys.exit(f"Usage: {sys.argv[0]} <jobname>")

    jobname = sys.argv[1]
    project_root = Path.cwd()
    
    # setting paths and making aux dir
    aux_file_path = project_root / f"{jobname}.scores.aux"
    output_dir = project_root / f"{jobname}-notaaux"
    output_dir.mkdir(exist_ok=True)

    if not aux_file_path.is_file():
        print(f"Auxiliary file {aux_file_path} not found. Run LaTeX first.")
        sys.exit(0)

    print(f"--- Reading score requests from {aux_file_path.name}")
    with open(aux_file_path, 'r') as f:
        score_requests = f.readlines()

    for request in score_requests:
        request = request.strip()
        if not request:
            continue

        try:
            parts = request.split('|', 10)
            if len(parts) != 11:
                print(f"  -> Warning: Malformed line in .scores.aux, skipping: {request}")
                continue
            score_id, score_type, file_path_str, font, unit, paperwidth_str, paperheight_str, topmargin_str, bottommargin_str, oddsidemargin_str, evensidemargin_str = parts
        except ValueError:
            print(f"  -> Warning: Malformed line in .scores.aux, skipping: {request}")
            continue

        print(f"- Processing score {score_id}: type '{score_type}', path '{file_path_str}'")

        input_file = project_root / file_path_str
        
        # Parse dimensions from LaTeX
        paperwidth_pt = parse_latex_dimension(paperwidth_str)
        paperheight_pt = parse_latex_dimension(paperheight_str)
        topmargin_pt = parse_latex_dimension(topmargin_str)
        bottommargin_pt = parse_latex_dimension(bottommargin_str)
        oddsidemargin_pt = parse_latex_dimension(oddsidemargin_str)
        evensidemargin_pt = parse_latex_dimension(evensidemargin_str)

        # Convert all relevant dimensions to millimeters for Verovio
        paperwidth_mm = pt_to_mm(paperwidth_pt)
        paperheight_mm = pt_to_mm(paperheight_pt)
        topmargin_mm = pt_to_mm(topmargin_pt)
        bottommargin_mm = pt_to_mm(bottommargin_pt)
        oddsidemargin_mm = pt_to_mm(oddsidemargin_pt)
        evensidemargin_mm = pt_to_mm(evensidemargin_pt)

        # Construct Verovio options dynamically
        verovio_options = {
            # "scaleToPageSize": True,
            # "scale": 100, # Default scale
            "mmOutput": True, # Output dimensions in mm for PDF readiness
            "footer": "none",
            "unit": unit,
            "font" : font
        }

        # Set 'breaks' option based on score type
        
        if score_type == 'inline':

            verovio_options["header"] = "none"

            verovio_options["pageHeight"] = 120
            verovio_options["pageWidth"] = (paperwidth_mm-oddsidemargin_mm-evensidemargin_mm+10) * 10 
            verovio_options["adjustPageHeight"] = True
            
            verovio_options["pageMarginTop"] = 0
            verovio_options["pageMarginBottom"] = 0

        elif score_type == 'fullscore':

            verovio_options["pageWidth"] = (paperwidth_mm-oddsidemargin_mm-evensidemargin_mm+10) * 10 
            verovio_options["pageHeight"] = (paperheight_mm-topmargin_mm-bottommargin_mm) * 10
            
            verovio_options["justifyVertically"] = True
            verovio_options["adjustPageHeight"] = True

        elif score_type == 'example':

            verovio_options["header"] = "none"
            verovio_options["breaks"] = "encoded"

            verovio_options["pageWidth"] = paperwidth_mm * 10 

            verovio_options["pageMarginLeft"] = oddsidemargin_mm * 10
            verovio_options["pageMarginRight"] = evensidemargin_mm * 10
            
            verovio_options["adjustPageHeight"] = True
            verovio_options["adjustPageWidth"] = True

        # Define output paths based on the unique score ID
        output_pdf = output_dir / f"nota-score-{score_id}.pdf"
        output_tex = output_dir / f"nota-score-{score_id}.tex"

        page_count = render_score(input_file, output_pdf, verovio_options)

        if page_count > 0:
            
            print(f"  -> Rendered {output_pdf.name}")
            
            # Write the .tex include file
            relative_pdf_path = os.path.relpath(output_pdf, output_dir)
            
            with open(output_tex, 'w') as f:
                if score_type == 'fullscore':
                    for i in range(1, page_count + 1):
                        f.write(f"\\makebox[\\textwidth][c]{{\\includegraphics[page={i}, scale=1]{{{relative_pdf_path}}}}}\\par\n")
                elif score_type == 'inline':
                    for i in range(1, page_count + 1):
                        f.write(f"\\makebox[\\textwidth][c]{{\\includegraphics[page={i}, scale=1]{{{relative_pdf_path}}}}}\\par\n")
                elif score_type == 'example':
                    f.write(f"\\makebox[\\textwidth][c]{{\\includegraphics[scale=1]{{{relative_pdf_path}}}}}\n")
                else:
                    for i in range(1, page_count + 1):
                        f.write(f"\\makebox[\\textwidth][c]{{\\includegraphics[page={i}, scale=1]{{{relative_pdf_path}}}}}\\par\n")
            print(f"  -> Created {output_tex.name}")


if __name__ == "__main__":
    main()