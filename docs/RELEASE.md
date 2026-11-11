# Release instructions

To release a new version, in a branch:

- populate the CHANGELOG.md file
- replace all occurrences of `0.0.1`
- ship the branch

Then create the release fully automatically:

```sh
git sync --all && git checkout main && git tag v0.0.1 && git push --tags
```

## Debug

Run `make release-snapshot` to try the release build locally.
