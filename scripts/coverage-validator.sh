#!/bin/bash

# Coverage Validator Script
# This script validates coverage against thresholds and enforces coverage policies

set -e

# Configuration
COVERAGE_THRESHOLD=${COVERAGE_THRESHOLD:-85}
BASELINE_FILE=${BASELINE_FILE:-".github/coverage-baseline.json"}
STRICT_MODE=${STRICT_MODE:-true}
COVERAGE_FILE=${COVERAGE_FILE:-"coverage.out"}
OUTPUT_DIR=${OUTPUT_DIR:-"coverage-validation"}

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

# Function to check dependencies
check_dependencies() {
    log_info "Checking dependencies..."
    
    local missing_tools=()
    
    # Check for required tools
    local required_tools=("go" "jq" "bc")
    for tool in "${required_tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            missing_tools+=("$tool")
        fi
    done
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        exit 1
    fi
    
    log_success "All dependencies are available"
}

# Function to validate coverage file
validate_coverage_file() {
    local coverage_file="$1"
    
    log_info "Validating coverage file: $coverage_file"
    
    if [ ! -f "$coverage_file" ]; then
        log_error "Coverage file not found: $coverage_file"
        return 1
    fi
    
    if [ ! -s "$coverage_file" ]; then
        log_error "Coverage file is empty: $coverage_file"
        return 1
    fi
    
    # Check if it's a valid Go coverage file
    if ! head -1 "$coverage_file" | grep -q "mode:"; then
        log_error "Invalid Go coverage file format: $coverage_file"
        return 1
    fi
    
    log_success "Coverage file is valid"
    return 0
}

# Function to calculate coverage percentage
calculate_coverage() {
    local coverage_file="$1"
    
    log_info "Calculating coverage percentage..."
    
    local coverage_percentage
    coverage_percentage=$(go tool cover -func="$coverage_file" | grep total | awk '{print $3}' | sed 's/%//')
    
    if [ -z "$coverage_percentage" ]; then
        log_error "Failed to calculate coverage percentage"
        return 1
    fi
    
    echo "$coverage_percentage"
}

# Function to load baseline coverage
load_baseline() {
    log_info "Loading coverage baseline..."
    
    if [ ! -f "$BASELINE_FILE" ]; then
        log_warning "Baseline file not found: $BASELINE_FILE"
        echo "0"
        return 0
    fi
    
    local baseline_coverage
    baseline_coverage=$(jq -r '.coverage' "$BASELINE_FILE" 2>/dev/null || echo "0")
    
    if [ "$baseline_coverage" = "null" ] || [ -z "$baseline_coverage" ]; then
        log_warning "Invalid baseline coverage, using 0"
        echo "0"
    else
        log_info "Baseline coverage: ${baseline_coverage}%"
        echo "$baseline_coverage"
    fi
}

# Function to validate against threshold
validate_threshold() {
    local current_coverage="$1"
    local threshold="$2"
    
    log_info "Validating coverage against threshold..."
    log_info "Current coverage: ${current_coverage}%"
    log_info "Required threshold: ${threshold}%"
    
    if (( $(echo "$current_coverage >= $threshold" | bc -l) )); then
        log_success "✅ Coverage validation PASSED: ${current_coverage}% >= ${threshold}%"
        return 0
    else
        log_error "❌ Coverage validation FAILED: ${current_coverage}% < ${threshold}%"
        return 1
    fi
}

# Function to compare with baseline
compare_baseline() {
    local current_coverage="$1"
    local baseline_coverage="$2"
    
    log_info "Comparing with baseline coverage..."
    
    if [ "$baseline_coverage" = "0" ]; then
        log_info "📊 Initial coverage measurement: ${current_coverage}%"
        echo "initial"
        return 0
    fi
    
    local coverage_change
    coverage_change=$(echo "$current_coverage - $baseline_coverage" | bc -l)
    
    if (( $(echo "$coverage_change >= 0" | bc -l) )); then
        log_success "📈 Coverage increased by ${coverage_change}%"
        echo "increased"
    else
        local coverage_change_abs
        coverage_change_abs=$(echo "$coverage_change" | sed 's/-//')
        log_warning "📉 Coverage decreased by ${coverage_change_abs}%"
        echo "decreased"
    fi
    
    return 0
}

# Function to generate detailed analysis
generate_analysis() {
    local coverage_file="$1"
    local current_coverage="$2"
    local baseline_coverage="$3"
    local threshold="$4"
    
    log_info "Generating detailed coverage analysis..."
    
    mkdir -p "$OUTPUT_DIR/analysis"
    
    # Generate function coverage report
    go tool cover -func="$coverage_file" > "$OUTPUT_DIR/analysis/function-coverage.txt"
    
    # Count functions by coverage level
    local total_functions
    local covered_functions
    local excellent_functions
    local good_functions
    local poor_functions
    local uncovered_functions
    
    total_functions=$(grep -v "total:" "$OUTPUT_DIR/analysis/function-coverage.txt" | wc -l)
    covered_functions=$(awk '$3 != "total:" && $3 != "0.0%" {count++} END {print count+0}' "$OUTPUT_DIR/analysis/function-coverage.txt")
    excellent_functions=$(awk '$3 != "total:" && $3 >= "90.0%" {count++} END {print count+0}' "$OUTPUT_DIR/analysis/function-coverage.txt")
    good_functions=$(awk '$3 != "total:" && $3 >= "70.0%" && $3 < "90.0%" {count++} END {print count+0}' "$OUTPUT_DIR/analysis/function-coverage.txt")
    poor_functions=$(awk '$3 != "total:" && $3 > "0.0%" && $3 < "70.0%" {count++} END {print count+0}' "$OUTPUT_DIR/analysis/function-coverage.txt")
    uncovered_functions=$(awk '$3 == "0.0%" {count++} END {print count+0}' "$OUTPUT_DIR/analysis/function-coverage.txt")
    
    # Create analysis report
    cat > "$OUTPUT_DIR/analysis/coverage-analysis.md" <<EOF
# Coverage Analysis Report

Generated on: $(date -u +"%Y-%m-%d %H:%M:%S UTC")

## Summary
- **Current Coverage**: ${current_coverage}%
- **Baseline Coverage**: ${baseline_coverage}%
- **Required Threshold**: ${threshold}%
- **Coverage Change**: $(echo "$current_coverage - $baseline_coverage" | bc -l)%

## Function Analysis
- **Total Functions**: $total_functions
- **Covered Functions**: $covered_functions ($(( covered_functions * 100 / total_functions ))%)
- **Excellent Coverage (≥90%)**: $excellent_functions
- **Good Coverage (70-89%)**: $good_functions  
- **Poor Coverage (1-69%)**: $poor_functions
- **No Coverage (0%)**: $uncovered_functions

## Coverage Distribution
EOF
    
    # Add coverage distribution chart (text-based)
    local excellent_bar
    local good_bar
    local poor_bar
    local uncovered_bar
    
    if [ "$total_functions" -gt 0 ]; then
        excellent_bar=$(printf "%*s" $(( excellent_functions * 50 / total_functions )) "" | tr ' ' '█')
        good_bar=$(printf "%*s" $(( good_functions * 50 / total_functions )) "" | tr ' ' '█')
        poor_bar=$(printf "%*s" $(( poor_functions * 50 / total_functions )) "" | tr ' ' '█')
        uncovered_bar=$(printf "%*s" $(( uncovered_functions * 50 / total_functions )) "" | tr ' ' '█')
    fi
    
    cat >> "$OUTPUT_DIR/analysis/coverage-analysis.md" <<EOF

\`\`\`
Excellent (≥90%): $excellent_bar ($excellent_functions)
Good (70-89%):    $good_bar ($good_functions)
Poor (1-69%):     $poor_bar ($poor_functions)
None (0%):        $uncovered_bar ($uncovered_functions)
\`\`\`

## Functions Requiring Attention

### Uncovered Functions (0% coverage)
EOF
    
    # List uncovered functions
    awk '$3 == "0.0%" {printf "- %s (%s)\n", $2, $1}' "$OUTPUT_DIR/analysis/function-coverage.txt" >> "$OUTPUT_DIR/analysis/coverage-analysis.md"
    
    cat >> "$OUTPUT_DIR/analysis/coverage-analysis.md" <<EOF

### Low Coverage Functions (< 70% coverage)
EOF
    
    # List low coverage functions
    awk '$3 != "total:" && $3 > "0.0%" && $3 < "70.0%" {printf "- %s (%s): %.1f%%\n", $2, $1, $3}' "$OUTPUT_DIR/analysis/function-coverage.txt" >> "$OUTPUT_DIR/analysis/coverage-analysis.md"
    
    log_success "Coverage analysis generated: $OUTPUT_DIR/analysis/coverage-analysis.md"
}

# Function to generate recommendations
generate_recommendations() {
    local current_coverage="$1"
    local threshold="$2"
    local validation_result="$3"
    
    log_info "Generating coverage recommendations..."
    
    mkdir -p "$OUTPUT_DIR/recommendations"
    
    cat > "$OUTPUT_DIR/recommendations/recommendations.md" <<EOF
# Coverage Improvement Recommendations

Generated on: $(date -u +"%Y-%m-%d %H:%M:%S UTC")

## Current Status
- **Coverage**: ${current_coverage}%
- **Threshold**: ${threshold}%
- **Gap**: $(echo "$threshold - $current_coverage" | bc -l)%
- **Status**: $(if [ "$validation_result" -eq 0 ]; then echo "✅ PASSED"; else echo "❌ FAILED"; fi)

EOF
    
    if [ "$validation_result" -ne 0 ]; then
        cat >> "$OUTPUT_DIR/recommendations/recommendations.md" <<EOF
## 🎯 Priority Actions

### Immediate (High Priority)
1. **Add Unit Tests**: Focus on uncovered functions identified in the analysis
2. **Test Critical Paths**: Prioritize testing for core business logic
3. **Mock Dependencies**: Use mocking to isolate units under test

### Short-term (Medium Priority)
1. **Integration Tests**: Add tests for component interactions
2. **Edge Cases**: Test error conditions and boundary cases
3. **Refactor for Testability**: Break down complex functions

### Long-term (Low Priority)
1. **Test Coverage Monitoring**: Set up automated coverage tracking
2. **Coverage Goals**: Establish incremental coverage targets
3. **Team Training**: Improve team's testing practices

## 📊 Coverage Targets

### Incremental Goals
- **Next milestone**: $(echo "$current_coverage + 10" | bc -l)%
- **Target**: ${threshold}%
- **Stretch goal**: $(echo "$threshold + 5" | bc -l)%

### Focus Areas
Based on the analysis, focus testing efforts on:
- Functions with 0% coverage (highest impact)
- Core business logic functions
- Error handling and edge cases
- Integration points between components

## 🛠️ Testing Strategies

### Unit Testing
- Use table-driven tests for multiple scenarios
- Mock external dependencies
- Test both success and failure paths
- Aim for 90%+ coverage on core functions

### Integration Testing
- Test component interactions
- Use test containers for external dependencies
- Validate end-to-end workflows
- Focus on critical user journeys

### Best Practices
- Write tests before fixing bugs (TDD approach)
- Keep tests simple and focused
- Use descriptive test names
- Maintain test code quality
EOF
    else
        cat >> "$OUTPUT_DIR/recommendations/recommendations.md" <<EOF
## ✅ Maintenance Recommendations

### Continue Good Practices
1. **Maintain Coverage**: Keep coverage above ${threshold}%
2. **Monitor Trends**: Watch for coverage regressions
3. **Review New Code**: Ensure new features include tests

### Improvement Opportunities
1. **Increase Coverage**: Target $(echo "$threshold + 5" | bc -l)%+ coverage
2. **Test Quality**: Focus on test effectiveness, not just coverage
3. **Performance Tests**: Add performance and load testing

### Advanced Testing
1. **Property-Based Testing**: Consider using property-based test frameworks
2. **Mutation Testing**: Validate test quality with mutation testing
3. **Coverage Analysis**: Regular deep-dive coverage analysis
EOF
    fi
    
    log_success "Recommendations generated: $OUTPUT_DIR/recommendations/recommendations.md"
}

# Function to update baseline
update_baseline() {
    local current_coverage="$1"
    local commit_sha="$2"
    local branch="$3"
    
    log_info "Updating coverage baseline..."
    
    # Create baseline directory if it doesn't exist
    mkdir -p "$(dirname "$BASELINE_FILE")"
    
    # Create new baseline
    cat > "$BASELINE_FILE" <<EOF
{
  "coverage": $current_coverage,
  "date": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "commit": "${commit_sha:-unknown}",
  "branch": "${branch:-unknown}",
  "threshold": $COVERAGE_THRESHOLD,
  "updated_by": "coverage-validator-script"
}
EOF
    
    log_success "Baseline updated to ${current_coverage}%"
}

# Function to enforce coverage policy
enforce_policy() {
    local validation_result="$1"
    local current_coverage="$2"
    local threshold="$3"
    
    log_info "Enforcing coverage policy..."
    
    if [ "$STRICT_MODE" = "true" ] && [ "$validation_result" -ne 0 ]; then
        log_error "🚨 COVERAGE POLICY VIOLATION"
        log_error "Current coverage (${current_coverage}%) is below the required threshold (${threshold}%)"
        log_error ""
        log_error "This build is blocked due to insufficient test coverage."
        log_error "Please add tests to improve coverage before proceeding."
        log_error ""
        log_error "See the generated analysis and recommendations for guidance."
        return 1
    elif [ "$validation_result" -ne 0 ]; then
        log_warning "⚠️ Coverage below threshold, but strict mode is disabled"
        log_warning "Consider adding tests to improve coverage"
        return 0
    else
        log_success "✅ Coverage policy compliance verified"
        return 0
    fi
}

# Function to create summary
create_summary() {
    local current_coverage="$1"
    local baseline_coverage="$2"
    local threshold="$3"
    local validation_result="$4"
    local comparison_result="$5"
    
    log_info "Creating coverage summary..."
    
    mkdir -p "$OUTPUT_DIR"
    
    cat > "$OUTPUT_DIR/coverage-summary.json" <<EOF
{
  "current_coverage": $current_coverage,
  "baseline_coverage": $baseline_coverage,
  "threshold": $threshold,
  "validation_passed": $(if [ "$validation_result" -eq 0 ]; then echo "true"; else echo "false"; fi),
  "comparison": "$comparison_result",
  "coverage_change": $(echo "$current_coverage - $baseline_coverage" | bc -l),
  "gap_to_threshold": $(echo "$threshold - $current_coverage" | bc -l),
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "strict_mode": $STRICT_MODE
}
EOF
    
    # Create markdown summary
    cat > "$OUTPUT_DIR/coverage-summary.md" <<EOF
# Coverage Validation Summary

## Results
- **Current Coverage**: ${current_coverage}%
- **Baseline Coverage**: ${baseline_coverage}%
- **Required Threshold**: ${threshold}%
- **Validation**: $(if [ "$validation_result" -eq 0 ]; then echo "✅ PASSED"; else echo "❌ FAILED"; fi)

## Analysis
- **Coverage Change**: $(echo "$current_coverage - $baseline_coverage" | bc -l)%
- **Gap to Threshold**: $(echo "$threshold - $current_coverage" | bc -l)%
- **Trend**: $comparison_result

## Generated Artifacts
- 📊 Detailed analysis: \`$OUTPUT_DIR/analysis/coverage-analysis.md\`
- 📋 Recommendations: \`$OUTPUT_DIR/recommendations/recommendations.md\`
- 📈 Function coverage: \`$OUTPUT_DIR/analysis/function-coverage.txt\`

Generated on: $(date -u +"%Y-%m-%d %H:%M:%S UTC")
EOF
    
    log_success "Coverage summary created"
}

# Main function
main() {
    log_info "Starting coverage validation..."
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -f|--file)
                COVERAGE_FILE="$2"
                shift 2
                ;;
            -t|--threshold)
                COVERAGE_THRESHOLD="$2"
                shift 2
                ;;
            -b|--baseline)
                BASELINE_FILE="$2"
                shift 2
                ;;
            -o|--output)
                OUTPUT_DIR="$2"
                shift 2
                ;;
            --strict)
                STRICT_MODE=true
                shift
                ;;
            --no-strict)
                STRICT_MODE=false
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # Check dependencies
    check_dependencies
    
    # Validate coverage file
    if ! validate_coverage_file "$COVERAGE_FILE"; then
        exit 1
    fi
    
    # Calculate current coverage
    local current_coverage
    current_coverage=$(calculate_coverage "$COVERAGE_FILE")
    
    # Load baseline coverage
    local baseline_coverage
    baseline_coverage=$(load_baseline)
    
    # Validate against threshold
    local validation_result=0
    if ! validate_threshold "$current_coverage" "$COVERAGE_THRESHOLD"; then
        validation_result=1
    fi
    
    # Compare with baseline
    local comparison_result
    comparison_result=$(compare_baseline "$current_coverage" "$baseline_coverage")
    
    # Generate analysis and recommendations
    generate_analysis "$COVERAGE_FILE" "$current_coverage" "$baseline_coverage" "$COVERAGE_THRESHOLD"
    generate_recommendations "$current_coverage" "$COVERAGE_THRESHOLD" "$validation_result"
    
    # Create summary
    create_summary "$current_coverage" "$baseline_coverage" "$COVERAGE_THRESHOLD" "$validation_result" "$comparison_result"
    
    # Update baseline if validation passed and we're on main branch
    if [ "$validation_result" -eq 0 ] && [ "${GITHUB_REF:-}" = "refs/heads/main" ]; then
        update_baseline "$current_coverage" "${GITHUB_SHA:-}" "${GITHUB_REF_NAME:-main}"
    fi
    
    # Enforce policy
    if ! enforce_policy "$validation_result" "$current_coverage" "$COVERAGE_THRESHOLD"; then
        exit 1
    fi
    
    log_success "Coverage validation completed successfully"
    log_info "Results available in: $OUTPUT_DIR"
}

# Help function
show_help() {
    cat <<EOF
Coverage Validator Script

Usage: $0 [OPTIONS]

OPTIONS:
    -f, --file FILE         Coverage file to validate (default: coverage.out)
    -t, --threshold NUM     Coverage threshold percentage (default: 85)
    -b, --baseline FILE     Baseline file path (default: .github/coverage-baseline.json)
    -o, --output DIR        Output directory (default: coverage-validation)
    --strict                Enable strict mode (fail on threshold violation)
    --no-strict             Disable strict mode
    -h, --help              Show this help message

ENVIRONMENT VARIABLES:
    COVERAGE_THRESHOLD      Coverage threshold percentage
    BASELINE_FILE           Baseline file path
    STRICT_MODE             Enable/disable strict mode (true/false)
    COVERAGE_FILE           Coverage file to validate
    OUTPUT_DIR              Output directory

EXAMPLES:
    $0                                          # Validate with defaults
    $0 -f coverage.out -t 90                   # Custom file and threshold
    $0 --strict -o validation-results          # Strict mode with custom output
    $0 --no-strict -b custom-baseline.json     # Custom baseline, non-strict

EOF
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi