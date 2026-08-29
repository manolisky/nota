#!/usr/bin/env bash

#
# Image Resize Script
# 
# Resizes an image by a specified percentage of its original dimensions
# Creates output file with "_resized_<percentage>" suffix
#
# Requirements: ImageMagick 7+ (magick command)
# 
# Examples:
#   ./resize_image.sh photo.jpg 50      # resize to 50% (half width, half height)
#   ./resize_image.sh image.png 25      # resize to 25% (quarter size)
#   ./resize_image.sh screenshot.webp 200  # resize to 200% (double size)
#

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Usage function
usage() {
    cat << EOF
Usage: $(basename "$0") <input_image> <resize_percentage>

Resizes an image by the specified percentage of its original dimensions.
Creates output file with "_resized_<percentage>" suffix.

Arguments:
  input_image        Path to input image file
  resize_percentage  Resize percentage (e.g., 50 for half size, 200 for double size)

Supported formats:
  JPEG/JPG, PNG, WEBP, TIFF, BMP, GIF

Examples:
  $(basename "$0") photo.jpg 50       # 50% size (half width, half height)
  $(basename "$0") image.png 25       # 25% size (quarter area)
  $(basename "$0") screenshot.webp 200 # 200% size (double width, double height)
  $(basename "$0") diagram.tiff 75    # 75% size

Notes:
  - 50% means half the width AND half the height (resulting in 25% of original area)
  - Original aspect ratio is preserved
  - Quality is maintained during resize

Requirements:
  ImageMagick 7+ (magick command must be in PATH)

EOF
}

# Error function
error() {
    echo -e "${RED}Error: $1${NC}" >&2
    usage
    exit 1
}

# Info function  
info() {
    echo -e "${GREEN}$1${NC}"
}

# Warning function
warn() {
    echo -e "${YELLOW}Warning: $1${NC}"
}

# Check ImageMagick availability
check_imagemagick() {
    if ! command -v magick >/dev/null 2>&1; then
        error "ImageMagick (magick command) not found. Please install ImageMagick 7+."
    fi
}

# Convert file size to MB
bytes_to_mb() {
    local bytes=$1
    echo "scale=2; $bytes / 1048576" | bc -l
}

# Get file size in bytes (cross-platform)
get_file_size() {
    local file=$1
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        stat -f%z "$file"
    else
        # Linux
        stat -c%s "$file"
    fi
}

# Get image dimensions
get_image_dimensions() {
    local file=$1
    magick identify -ping -format "%w %h" "$file" 2>/dev/null
}

# Validate arguments
validate_arguments() {
    # Check argument count
    if [[ $# -ne 2 ]]; then
        error "Exactly 2 arguments required, got $#"
    fi

    local input_file="$1"
    local resize_percentage="$2"

    # Check if input file exists and is readable
    if [[ ! -f "$input_file" ]]; then
        echo -e "${RED}Error: Input file '$input_file' does not exist. Ensure the file is present or created by a previous step.${NC}" >&2
        exit 1
    fi

    if [[ ! -r "$input_file" ]]; then
        echo -e "${RED}Error: Input file '$input_file' is not readable${NC}" >&2
        exit 1
    fi

    # Check file extension
    local extension="${input_file##*.}"
    extension=$(echo "$extension" | tr '[:upper:]' '[:lower:]')
    
    case "$extension" in
        jpg|jpeg|png|webp|tiff|tif|bmp|gif)
            # Valid formats
            ;;
        *)
            error "Unsupported file format: .$extension. Supported: jpg, jpeg, png, webp, tiff, bmp, gif"
            ;;
    esac

    # Validate resize percentage (positive number)
    if ! [[ "$resize_percentage" =~ ^[0-9]*\.?[0-9]+$ ]] || (( $(echo "$resize_percentage <= 0" | bc -l) )); then
        error "Resize percentage must be a positive number (got: '$resize_percentage')"
    fi
}

# Generate output filename
generate_output_filename() {
    local input_file="$1"
    local resize_percentage="$2"
    
    local dir=$(dirname "$input_file")
    local basename=$(basename "$input_file")
    local filename_no_ext="${basename%.*}"
    local extension="${basename##*.}"
    
    # Clean percentage for filename (remove trailing zeros after decimal)
    local clean_percentage=$(echo "$resize_percentage" | sed 's/\.0*$//' | sed 's/\.$//')
    
    echo "${dir}/${filename_no_ext}_resized_${clean_percentage}.${extension}"
}

# Resize image
resize_image() {
    local input_file="$1"
    local resize_percentage="$2"
    local output_file="$3"
    
    info "Resizing $(basename "$input_file") to ${resize_percentage}%..."
    
    # Get original dimensions
    read -r original_width original_height <<< $(get_image_dimensions "$input_file")
    
    # Calculate new dimensions
    local new_width=$(echo "scale=0; $original_width * $resize_percentage / 100" | bc -l | cut -d. -f1)
    local new_height=$(echo "scale=0; $original_height * $resize_percentage / 100" | bc -l | cut -d. -f1)
    
    info "  Original dimensions: ${original_width}x${original_height}"
    info "  New dimensions: ${new_width}x${new_height}"
    
    # Perform the resize
    if magick "$input_file" -resize "${new_width}x${new_height}" "$output_file" 2>/dev/null; then
        info "✓ Resize completed successfully!"
        return 0
    else
        error "Failed to resize image"
    fi
}

# Main function
main() {
    # Check dependencies
    check_imagemagick
    
    # Validate arguments
    validate_arguments "$@"
    
    local input_file="$1"
    local resize_percentage="$2"
    
    # Generate output filename
    local output_file
    output_file=$(generate_output_filename "$input_file" "$resize_percentage")
    
    # Check if output file already exists
    if [[ -f "$output_file" ]]; then
        warn "Output file '$output_file' already exists and will be overwritten"
    fi
    
    # Display initial file info
    local original_size=$(get_file_size "$input_file")
    info "Input file: $(basename "$input_file") ($(bytes_to_mb "$original_size")MB)"
    info "Resize to: ${resize_percentage}%"
    info "Output file: $(basename "$output_file")"
    echo
    
    # Perform resize
    resize_image "$input_file" "$resize_percentage" "$output_file"
    
    # Display results
    local final_size=$(get_file_size "$output_file")
    local size_reduction=$(echo "scale=1; (($original_size - $final_size) * 100) / $original_size" | bc -l)
    
    echo
    info "✓ Operation completed!"
    info "  Original: $(bytes_to_mb "$original_size")MB"
    info "  Resized: $(bytes_to_mb "$final_size")MB"
    if (( $(echo "$final_size < $original_size" | bc -l) )); then
        info "  File size reduction: ${size_reduction}%"
    else
        info "  File size increase: $(echo "scale=1; (($final_size - $original_size) * 100) / $original_size" | bc -l)%"
    fi
    info "  Output: $output_file"
}

# Run main function with all arguments
main "$@"
