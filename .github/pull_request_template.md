## Description

<!-- Provide a brief description of the changes in this PR -->

## Type of Change

<!-- Check the relevant option -->
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update
- [ ] New update script (adding support for a new tool)
- [ ] Performance improvement
- [ ] Code refactoring

## Related Issue

<!-- Link to the issue this PR addresses, if applicable -->
Fixes #<!-- issue number -->

## Changes Made

<!-- List the specific changes made in this PR -->
-
-
-

## Testing

<!-- Describe how you tested these changes -->

**Test Configuration**:
- OS: <!-- e.g., macOS 14.0, Ubuntu 22.04 -->
- Go version: <!-- e.g., 1.25.4 -->

**Tests performed**:
- [ ] `go build` succeeds
- [ ] `go test -v ./...` passes
- [ ] Tested manually with `update-ng --only <new-tool>`
- [ ] Tested with `goreleaser release --snapshot --clean --skip=publish`
- [ ] Verified script works correctly
- [ ] Tested on target platform (if platform-specific)

## Checklist

- [ ] My code follows the project's code style
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have updated the documentation accordingly
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] Any dependent changes have been merged and published

## For New Update Scripts

<!-- If adding a new update script, complete this section -->

- [ ] Script follows naming convention: `update-<toolname>`
- [ ] Script is executable: `chmod +x bin/update-<toolname>`
- [ ] Script has proper shebang: `#!/bin/sh`
- [ ] Added description to `getScriptDescriptions()` in `main.go`
- [ ] Added to appropriate category in `scriptCategories` map
- [ ] Script handles tool not installed gracefully
- [ ] Script provides meaningful output
- [ ] Tested script on target platform

**Tool Name**: <!-- e.g., terraform, ansible -->
**Category**: <!-- e.g., Cloud & Infrastructure, Development Tools -->
**Description**: <!-- Brief description of what the tool does -->

## Screenshots (if applicable)

<!-- Add screenshots to help explain your changes -->

## Additional Notes

<!-- Any additional information that reviewers should know -->
