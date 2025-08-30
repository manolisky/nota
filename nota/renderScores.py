import sys
import os
import tempfile
import re
from pathlib import Path
import concurrent.futures

# This script is part of a two-pass build system.
# 1. LaTeX runs first, creating a .scores.aux file with a list of required scores.
# 2. This script reads that .scores.aux file and renders the scores.
# 3. LaTeX runs again to include the newly created score files.

# Import rendering libraries
import verovio
import cairosvg
from PyPDF2 import PdfMerger
import subprocess

# auxiliary function needed to render string value in pt from latex from aux file to int pixel for Verovio.
def parse_latex_dimension(dim_str: str) -> int:
    
    # match string in pt -> to float in pt
    match = re.match(r"([0-9.]+)", dim_str)
    
    # string in pt -> to float in pt
    dimension_in_pt = float(match.group(1))

    # float in pt -> to float in mm
    dimension_in_mm = dimension_in_pt * 0.3527
    
    # get float in mm -> to int in pixels
    dimension_in_pixels = int(dimension_in_mm * 10) 
    
    if dimension_in_pixels:
        return dimension_in_pixels
    return 0.0 # Default or error value

def convert_svg_to_pdf(svg: str, pdf_path: Path):
    """Helper function to convert SVG string to PDF file."""
    cairosvg.svg2pdf(bytestring=svg.encode('utf-8'), write_to=str(pdf_path))

def convert_page_to_pdf(args):
    i, svg, tmpdir = args
    pdf_page = Path(tmpdir) / f"p{i}.pdf"
    try:
        convert_svg_to_pdf(svg, pdf_page)
        return (i, str(pdf_page))
    except Exception as e:
        print(f"  -> Error converting page {i} to PDF: {e}")
        return (i, None)

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

    # convert svg to pdf pages in parallel to speed up cairoSVG conversion
    with tempfile.TemporaryDirectory() as tmpdir:
        svg_pages = []
        for i in range(1, page_count + 1):
            try:
                svg = tk.renderToSVG(i)
                svg_pages.append((i, svg))
            except Exception as e:
                print(f"  -> Error rendering page {i} of {input_file.name} to SVG: {e}")
                return 0

        # Prepare arguments for parallel conversion
        convert_args = [(i, svg, tmpdir) for i, svg in svg_pages]

        pages = []
        with concurrent.futures.ProcessPoolExecutor() as executor:
            results = executor.map(convert_page_to_pdf, convert_args)
            for i, pdf_path in results:
                if pdf_path is None:
                    return 0
                pages.append((i, pdf_path))

        # Sort pages by page number to ensure correct order before merging
        pages.sort(key=lambda x: x[0])

        # Merge Pages to single pdf write to output_pdf path
        merger = PdfMerger()
        for _, p in pages:
            merger.append(p)
        merger.write(str(output_pdf))
        merger.close()
    
    return page_count

def process_request(request: str, project_root: Path, output_dir: Path):
    request = request.strip()
    if not request:
        return None

    try:
        parts = request.split('|', 10)
        if len(parts) != 11:
            print(f"  -> Warning: Malformed line in .scores.aux, skipping: {request}")
            return None
        score_id, score_type, file_path_str, font, unit, paperwidth_str, paperheight_str, topmargin_str, bottommargin_str, oddsidemargin_str, evensidemargin_str = parts
    except ValueError:
        print(f"  -> Warning: Malformed line in .scores.aux, skipping: {request}")
        return None

    print(f"- Processing score {score_id}: type '{score_type}', path '{file_path_str}'")

    input_file = project_root / file_path_str

    # pixel values for Verovio
    paperwidth_px = parse_latex_dimension(paperwidth_str)
    paperheight_px = parse_latex_dimension(paperheight_str)
    topmargin_px = parse_latex_dimension(topmargin_str)
    bottommargin_px = parse_latex_dimension(bottommargin_str)
    oddsidemargin_px = parse_latex_dimension(oddsidemargin_str)
    evensidemargin_px = parse_latex_dimension(evensidemargin_str)

    # derive text width and height
    textwidth_px = paperwidth_px - oddsidemargin_px - evensidemargin_px
    textheight_px = paperheight_px - topmargin_px - bottommargin_px

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

        verovio_options["pageHeight"] = 100
        verovio_options["pageWidth"] = textwidth_px + 100

        verovio_options["adjustPageHeight"] = True
        
        verovio_options["pageMarginTop"] = 0
        verovio_options["pageMarginBottom"] = 0

    elif score_type == 'fullscore':

        verovio_options["pageWidth"] = textwidth_px + 100 
        verovio_options["pageHeight"] = textheight_px + 100

        verovio_options["pageMarginTop"] = 0
        verovio_options["pageMarginBottom"] = 120
        
        verovio_options["justifyVertically"] = True
        verovio_options["adjustPageHeight"] = True

    elif score_type == 'example':

        verovio_options["header"] = "none"
        verovio_options["breaks"] = "encoded"

        verovio_options["pageWidth"] = paperheight_px 

        verovio_options["pageMarginLeft"] = oddsidemargin_px
        verovio_options["pageMarginRight"] = evensidemargin_px
        
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
        return score_id
    return None

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

    # Use ProcessPoolExecutor to process score requests in parallel.
    # Each score request is submitted to the pool of worker processes.
    # We keep a mapping from each Future object to its corresponding request string,
    # so that when a Future completes, we know which request it corresponds to.
    with concurrent.futures.ProcessPoolExecutor() as executor:
        future_to_request = {executor.submit(process_request, request, project_root, output_dir): request for request in score_requests}
        
        # Iterate over futures as they complete.
        for future in concurrent.futures.as_completed(future_to_request):
            # Retrieve the original request string for this future.
            request = future_to_request[future]
            try:
                # future.result() returns the value returned by process_request.
                # This will be the score_id if processing was successful,
                # or None if the request was skipped or failed.
                result = future.result()
                if result is not None:
                    print(f"Finished processing score {result}")
                else:
                    print(f"Skipped or failed processing a request.")
            except Exception as exc:
                # If an exception was raised during processing, handle it here.
                print(f"Request generated an exception: {exc}")

if __name__ == "__main__":
    main()