#!/bin/bash

# Coverage Aggregator Script
# This script aggregates coverage data from multiple test runs and generates comprehensive reports

set -e

# Configuration
COVERAGE_THRESHOLD=${COVERAGE_THRESHOLD:-85}
OUTPUT_DIR=${OUTPUT_DIR:-"coverage-output"}
VERBOSE=${VERBOSE:-false}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if required tools are installed
check_dependencies() {
    log_info "Checking dependencies..."
    
    local missing_tools=()
    
    # Check for Go
    if ! command -v go &> /dev/null; then
        missing_tools+=("go")
    fi
    
    # Check for coverage tools
    local coverage_tools=("gocov" "gocov-xml" "gocovmerge" "gocover-cobertura")
    for tool in "${coverage_tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            missing_tools+=("$tool")
        fi
    done
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        log_info "Install missing tools with:"
        for tool in "${missing_tools[@]}"; do
            case $tool in
                "gocov")
                    echo "  go install github.com/axw/gocov/gocov@latest"
                    ;;
                "gocov-xml")
                    echo "  go install github.com/AlekSi/gocov-xml@latest"
                    ;;
                "gocovmerge")
                    echo "  go install github.com/wadey/gocovmerge@latest"
                    ;;
                "gocover-cobertura")
                    echo "  go install github.com/t-yuki/gocover-cobertura@latest"
                    ;;
            esac
        done
        exit 1
    fi
    
    log_success "All dependencies are available"
}

# Function to create output directories
setup_directories() {
    log_info "Setting up output directories..."
    
    mkdir -p "$OUTPUT_DIR"/{html,xml,json,badges,analysis}
    
    log_success "Output directories created"
}

# Function to find coverage files
find_coverage_files() {
    log_info "Finding coverage files..."
    
    local coverage_files=()
    
    # Look for coverage files in common locations
    local search_paths=(
        "coverage*.out"
        "*/coverage*.out"
        "coverage-data/*/coverage.out"
        "test-results/coverage*.out"
    )
    
    for pattern in "${search_paths[@]}"; do
        while IFS= read -r -d '' file; do
            if [ -s "$file" ]; then
                coverage_files+=("$file")
                if [ "$VERBOSE" = true ]; then
                    log_info "Found coverage file: $file"
                fi
            fi
        done < <(find . -name "$pattern" -type f -print0 2>/dev/null || true)
    done
    
    # Remove duplicates
    local unique_files=($(printf "%s\n" "${coverage_files[@]}" | sort -u))
    
    if [ ${#unique_files[@]} -eq 0 ]; then
        log_error "No coverage files found"
        exit 1
    fi
    
    log_success "Found ${#unique_files[@]} coverage files"
    printf '%s\n' "${unique_files[@]}"
}

# Function to merge coverage files
merge_coverage_files() {
    local coverage_files=("$@")
    local merged_file="$OUTPUT_DIR/coverage-merged.out"
    
    log_info "Merging ${#coverage_files[@]} coverage files..."
    
    if [ ${#coverage_files[@]} -eq 1 ]; then
        cp "${coverage_files[0]}" "$merged_file"
        log_info "Single coverage file copied"
    else
        gocovmerge "${coverage_files[@]}" > "$merged_file"
        log_success "Coverage files merged"
    fi
    
    echo "$merged_file"
}

# Function to generate HTML report
generate_html_report() {
    local coverage_file="$1"
    local html_file="$OUTPUT_DIR/html/coverage.html"
    
    log_info "Generating HTML coverage report..."
    
    go tool cover -html="$coverage_file" -o "$html_file"
    
    log_success "HTML report generated: $html_file"
}

# Function to generate XML reports
generate_xml_reports() {
    local coverage_file="$1"
    
    log_info "Generating XML coverage reports..."
    
    # Cobertura XML format
    local cobertura_file="$OUTPUT_DIR/xml/cobertura-coverage.xml"
    gocover-cobertura < "$coverage_file" > "$cobertura_file"
    log_success "Cobertura XML report generated: $cobertura_file"
    
    # Generic XML format
    local xml_file="$OUTPUT_DIR/xml/coverage.xml"
    gocov convert "$coverage_file" | gocov-xml > "$xml_file"
    log_success "Generic XML report generated: $xml_file"
}

# Function to generate JSON report
generate_json_report() {
    local coverage_file="$1"
    local json_file="$OUTPUT_DIR/json/coverage.json"
    
    log_info "Generating JSON coverage report..."
    
    gocov convert "$coverage_file" > "$json_file"
    
    log_success "JSON report generated: $json_file"
}

# Function to generate function coverage report
generate_function_report() {
    local coverage_file="$1"
    local func_file="$OUTPUT_DIR/coverage-func.txt"
    
    log_info "Generating function coverage report..."
    
    go tool cover -func="$coverage_file" > "$func_file"
    
    log_success "Function coverage report generated: $func_file"
}

# Function to calculate coverage metrics
calculate_metrics() {
    local coverage_file="$1"
    
    log_info "Calculating coverage metrics..."
    
    # Overall coverage
    local overall_coverage
    overall_coverage=$(go tool cover -func="$coverage_file" | grep total | awk '{print $3}' | sed 's/%//')
    
    # Function counts
    local total_functions
    local covered_functions
    total_functions=$(go tool cover -func="$coverage_file" | grep -v total | wc -l)
    covered_functions=$(go tool cover -func="$coverage_file" | grep -v total | awk '$3 != "0.0%" {count++} END {print count+0}')
    
    # Function coverage percentage
    local function_coverage=0
    if [ "$total_functions" -gt 0 ]; then
        function_coverage=$(( covered_functions * 100 / total_functions ))
    fi
    
    # Create metrics file
    local metrics_file="$OUTPUT_DIR/analysis/metrics.json"
    cat > "$metrics_file" <<EOF
{
  "overall_coverage": $overall_coverage,
  "total_functions": $total_functions,
  "covered_functions": $covered_functions,
  "function_coverage": $function_coverage,
  "threshold": $COVERAGE_THRESHOLD,
  "passes_threshold": $(if (( $(echo "$overall_coverage >= $COVERAGE_THRESHOLD" | bc -l) )); then echo "true"; else echo "false"; fi),
  "generated_at": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
}
EOF
    
    log_success "Coverage metrics calculated"
    log_info "Overall coverage: ${overall_coverage}%"
    log_info "Function coverage: ${function_coverage}% (${covered_functions}/${total_functions})"
    
    # Check threshold
    if (( $(echo "$overall_coverage >= $COVERAGE_THRESHOLD" | bc -l) )); then
        log_success "Coverage meets threshold (${COVERAGE_THRESHOLD}%)"
        return 0
    else
        log_error "Coverage below threshold: ${overall_coverage}% < ${COVERAGE_THRESHOLD}%"
        return 1
    fi
}

# Function to generate package analysis
generate_package_analysis() {
    local coverage_file="$1"
    local analysis_file="$OUTPUT_DIR/analysis/package-analysis.md"
    
    log_info "Generating package coverage analysis..."
    
    cat > "$analysis_file" <<EOF
# Package Coverage Analysis

Generated on: $(date)

## Package Coverage Breakdown

| Package | Coverage | Functions |
|---------|----------|-----------|
EOF
    
    # Parse function coverage for package breakdown
    awk '/^[^[:space:]]/ && !/^total:/ {
        package = $1
        gsub(/\/[^\/]*$/, "", package)
        if (package == "") package = "main"
        coverage[package] += $3
        count[package]++
    }
    END {
        for (pkg in coverage) {
            avg = coverage[pkg] / count[pkg]
            printf "| %s | %.1f%% | %d |\n", pkg, avg, count[pkg]
        }
    }' <(go tool cover -func="$coverage_file") >> "$analysis_file"
    
    cat >> "$analysis_file" <<EOF

## Low Coverage Functions (< 50%)

| Function | File | Coverage |
|----------|------|----------|
EOF
    
    awk '$3 != "total:" && $3 != "0.0%" && $3 < "50.0%" {
        gsub(/%/, "", $3)
        printf "| %s | %s | %.1f%% |\n", $2, $1, $3
    }' <(go tool cover -func="$coverage_file") >> "$analysis_file"
    
    cat >> "$analysis_file" <<EOF

## Uncovered Functions

| Function | File |
|----------|------|
EOF
    
    awk '$3 == "0.0%" {
        printf "| %s | %s |\n", $2, $1
    }' <(go tool cover -func="$coverage_file") >> "$analysis_file"
    
    log_success "Package analysis generated: $analysis_file"
}

# Function to generate coverage badge
generate_badge() {
    local coverage_percentage="$1"
    local badge_file="$OUTPUT_DIR/badges/coverage-badge.md"
    
    log_info "Generating coverage badge..."
    
    # Determine badge color based on coverage
    local color="red"
    if (( $(echo "$coverage_percentage >= 85" | bc -l) )); then
        color="brightgreen"
    elif (( $(echo "$coverage_percentage >= 70" | bc -l) )); then
        color="yellow"
    elif (( $(echo "$coverage_percentage >= 50" | bc -l) )); then
        color="orange"
    fi
    
    # Generate badge markdown
    local badge_url="https://img.shields.io/badge/coverage-${coverage_percentage}%25-${color}"
    echo "![Coverage](${badge_url})" > "$badge_file"
    
    # Also create a simple text file with the URL
    echo "$badge_url" > "$OUTPUT_DIR/badges/badge-url.txt"
    
    log_success "Coverage badge generated: $badge_file"
}

# Function to create summary report
create_summary() {
    local metrics_file="$OUTPUT_DIR/analysis/metrics.json"
    local summary_file="$OUTPUT_DIR/coverage-summary.md"
    
    log_info "Creating coverage summary..."
    
    if [ ! -f "$metrics_file" ]; then
        log_error "Metrics file not found: $metrics_file"
        return 1
    fi
    
    local overall_coverage
    local total_functions
    local covered_functions
    local function_coverage
    local passes_threshold
    
    overall_coverage=$(jq -r '.overall_coverage' "$metrics_file")
    total_functions=$(jq -r '.total_functions' "$metrics_file")
    covered_functions=$(jq -r '.covered_functions' "$metrics_file")
    function_coverage=$(jq -r '.function_coverage' "$metrics_file")
    passes_threshold=$(jq -r '.passes_threshold' "$metrics_file")
    
    cat > "$summary_file" <<EOF
# Coverage Analysis Summary

## Overall Results
- **Coverage**: ${overall_coverage}% (Threshold: ${COVERAGE_THRESHOLD}%)
- **Status**: $(if [ "$passes_threshold" = "true" ]; then echo "✅ PASSED"; else echo "❌ FAILED"; fi)
- **Functions**: ${covered_functions}/${total_functions} covered (${function_coverage}%)

## Generated Reports
- 📄 HTML Report: \`${OUTPUT_DIR}/html/coverage.html\`
- 📊 XML Report: \`${OUTPUT_DIR}/xml/coverage.xml\`
- 📋 JSON Report: \`${OUTPUT_DIR}/json/coverage.json\`
- 📦 Package Analysis: \`${OUTPUT_DIR}/analysis/package-analysis.md\`

## Badge
![Coverage](https://img.shields.io/badge/coverage-${overall_coverage}%25-$(if (( $(echo "$overall_coverage >= 85" | bc -l) )); then echo "brightgreen"; elif (( $(echo "$overall_coverage >= 70" | bc -l) )); then echo "yellow"; elif (( $(echo "$overall_coverage >= 50" | bc -l) )); then echo "orange"; else echo "red"; fi))

Generated on: $(date)
EOF
    
    log_success "Coverage summary created: $summary_file"
}

# Main function
main() {
    log_info "Starting coverage aggregation and analysis..."
    
    # Check dependencies
    check_dependencies
    
    # Setup directories
    setup_directories
    
    # Find coverage files
    local coverage_files
    mapfile -t coverage_files < <(find_coverage_files)
    
    # Merge coverage files
    local merged_file
    merged_file=$(merge_coverage_files "${coverage_files[@]}")
    
    # Generate reports
    generate_html_report "$merged_file"
    generate_xml_reports "$merged_file"
    generate_json_report "$merged_file"
    generate_function_report "$merged_file"
    
    # Calculate metrics and check threshold
    local threshold_passed=0
    if calculate_metrics "$merged_file"; then
        threshold_passed=1
    fi
    
    # Generate analysis
    generate_package_analysis "$merged_file"
    
    # Generate badge
    local coverage_percentage
    coverage_percentage=$(jq -r '.overall_coverage' "$OUTPUT_DIR/analysis/metrics.json")
    generate_badge "$coverage_percentage"
    
    # Create summary
    create_summary
    
    log_success "Coverage aggregation and analysis completed"
    log_info "Results available in: $OUTPUT_DIR"
    
    # Exit with appropriate code
    if [ $threshold_passed -eq 1 ]; then
        exit 0
    else
        exit 1
    fi
}

# Help function
show_help() {
    cat <<EOF
Coverage Aggregator Script

Usage: $0 [OPTIONS]

OPTIONS:
    -h, --help              Show this help message
    -t, --threshold NUM     Set coverage threshold (default: 85)
    -o, --output DIR        Set output directory (default: coverage-output)
    -v, --verbose           Enable verbose output

ENVIRONMENT VARIABLES:
    COVERAGE_THRESHOLD      Coverage threshold percentage (default: 85)
    OUTPUT_DIR              Output directory (default: coverage-output)
    VERBOSE                 Enable verbose output (true/false)

EXAMPLES:
    $0                      # Run with defaults
    $0 -t 90 -o reports     # Set threshold to 90% and output to reports/
    $0 --verbose            # Run with verbose output

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -t|--threshold)
            COVERAGE_THRESHOLD="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        *)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Run main function
main "$@"