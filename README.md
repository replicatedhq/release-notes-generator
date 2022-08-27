# Release Notes Generator

Takes the diff of two releases, finds all pull requests involved, and takes out their release notes and orders them into features/improvements/bugs.

### Usage

An example workflow is included at [./.github/workflows/main.yml](./.github/workflows/main.yml)

```yaml
name: Generate release notes for Replicated KOTS
on: [push]

jobs:
  test_action_job:
    runs-on: ubuntu-latest
    name: Generate release notes
    steps:
    - name: Checkout
      uses: actions/checkout@v1
    - id: test-action
      uses: replicatedhq/release-notes-generator@main
      with:
        owner-repo: replicatedhq/kots
        base: 'v1.81.1'
        head: 'v1.82.0'
        title: '1.82.0'
        description: 'Support for Kubernetes: 1.21, 1.22, 1.23, and 1.24'
        include-pr-links: false
        github-token: ${{ secrets.GITHUB_TOKEN }}
    - name: Print the output
      run: echo "${{ steps.test-action.outputs.release-notes }}"
```

### Inputs

#### owner-repo (required)

The owner/repo of the Github repository to be used.

#### base (optional)

The release tag to use as the base of the release notes diff.

#### head (required)

The release tag to use as the head of the release notes diff.

#### title

The release notes title.

#### description

Description to be added to the release notes.

#### include-pr-links

Include links back to pull requests.

#### github-token

Github API token to use to avoid rate limiting.

### Outputs

#### release-notes

The generated release notes markdown.
