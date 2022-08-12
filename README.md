# KOTS Release Helper

Takes the diff of the last release and main, finds all pull requests involved, and takes out their release notes and orders them into features/bugs.

## Running

*Note, Github API rate limiting may impact the ability to run this script.
To avoid this, set the environment variable GITHUB_AUTH_TOKEN with a token that has `repo:status` permissions.*

```sh
go run main.go
```

Release notes markdown is written to stdout. Example:

```
## 1.81.0

Released on August 12, 2022

Support for Kubernetes: 1.21, 1.22, 1.23, and 1.24

### New Features {#new-features-1-90-0}
* Adds support for the `alias` field in Helm chart dependencies.
* Adds support for image tags and digests to be used together for online installations.
* Changes the default value for `helmVersion` from `v2` to `v3` for the [HelmChart](/reference/custom-resource-helmchart) custom resource.

### Bug Fixes {#bug-fixes-1-90-0}
* (alpha) Fix bug where license tab would not show for helm managed applications.
* Fixes an issue that can cause `Namespace` manifests packaged in Helm charts to be excluded from deployment, causing namespaces to not be created when [useHelmInstall](/reference/custom-resource-helmchart#usehelminstall) is set to `true` and [namespace](/reference/custom-resource-helmchart#usehelminstall) is an empty string.
* Improves the UI responsiveness on the configuration page.
* Fixes an issue where GitOps was being enabled before the deploy key was added to the git provider.
* (alpha) hide copy command in UI when clipboard is not available.
```

## Configuration

```
  -base string
        Base of release notes diff (defaults to the last release)
  -head string
        Head of release notes diff (default "main")
  -pr-links
        Include links back to pull requests
  -semver string
        Override the automatically determined semver for the release
  -supported-versions string
        Comma-separated list of supported Kubernetes versions (default "1.21,1.22,1.23,1.24")
```

## Next Steps

Currently, there is no way for the script to differentiate between "New Features" and "Improvements". Potential solution could be to require sub-labels for `type::feature` PRs.