# KOTS Release Helper

Takes the latest nightly build, finds all pull requests involved, and takes out their release notes and orders them into features/bugs.

It also prints out the latest non pre-release build.

## Running

```sh
go run main.go
```

Should produce something like

```
Last release:  v1.66.0
features:
	 - ability to provide a strict preflight spec which should run and not fail before deploying an app.
- override flag `--skip-preflights=true` and run preflights when there are strict preflights present.
- disable `deploy button` when there is a strict preflights spec and if preflight is still running or strict preflight has failed.
- when flag `--deploy` is provided or deployment is triggered, if preflights are still running and has a strict preflight, poll until the preflights are finished running and analyze the preflight result for any failed strict preflights. If a strict preflight has failed, deployment is not continued and an appropriate error message is returned.
- if preflights are running and deployment is triggered, deployment will poll for preflights execution to finish with a timeout of 15 minutes.
	 - Application release channels that do not have [semantic versioning](https://docs.replicated.com/vendor/releases-understanding#semantic-versioning) enabled can now perform automatic release deployments. The most recent release will be used when updating, regardless of its version tag.
bugs:
	 - Fixes bug where the Cluster Management tab would not be initially present if the application is installed via kURL
	 Fixes an issue where trying to re-download a pending app version after upgrading from KOTS 1.65 would fail due to an invalid license for that version.
	 Fixes a bug where the app icon in the metadata would not show as the favicon on the TLS pages.
```