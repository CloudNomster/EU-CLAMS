# Release Process

This document outlines the process for creating new releases of EU-CLAMS.

## How to Create a New Release

1. Make sure all your changes are committed and tested in a feature branch.

2. Create a pull request targeting the `main` branch.

3. Once the pull request is approved and ready to merge, you'll need to:
   
   a. Update the version number in `FyneApp.toml` in your feature branch
   b. Commit this change to your feature branch
   c. Merge the pull request into `main`

4. After the PR is merged, create and push a new tag with the version number:

   ```powershell
   # Make sure you're on the main branch and up-to-date
   git checkout main
   git pull

   # Create a new tag (replace x.y.z with the actual version number)
   git tag v1.2.3
   
   # Push the tag to GitHub
   git push origin v1.2.3
   ```

5. The tag push will automatically trigger the release workflow (not regular pushes to main), which will:
   - Build the application
   - Create a ZIP package with the necessary files
   - Create a GitHub release with the executable
   - Generate release notes based on the commits since the last release

## Version Numbering

We follow semantic versioning (SemVer) for our releases:

- **Major version** (x.0.0): Significant changes, potentially breaking compatibility
- **Minor version** (0.x.0): New features, no breaking changes
- **Patch version** (0.0.x): Bug fixes and minor improvements

## Release Content

Each release includes:

1. The Windows executable (`eu-clams.exe`)
2. A ZIP package containing:
   - The executable
   - Any other necessary files (does not include configuration files)

## Release Notes

The release notes are automatically generated based on the commits since the last release. To ensure good quality release notes:

1. Write clear, descriptive commit messages
2. Reference issue numbers in commits where applicable
3. Use conventional commit prefixes when possible (feat:, fix:, docs:, etc.)
