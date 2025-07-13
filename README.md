# TNSEngineerEdition - WailsClient
This repository contains source code, and tests for client component of engineering thesis:  \
**Concurrent interactive simulator of Krakow's tram network**

## Development setup and notes
In order to efficiently and cleanly develop the application, we need to follow a few simple rules.

### Branch names
Branches created in this repository should be named in the following way:

`<author-names>/<short-description>`

The `<author-names>` field should correspond to your GitHub usernames. The names should be ordered alphabetically and be separated with `+` characters.

The `<short-description>` field should include a summary of the changes.

Examples of correct branch names:
```
RCRalph/added-pre-commit-check
Codefident+RCRalph/server-integration-tests
olobuszolo+Redor114/tram-network-graph-transformation
Codefident+olobuszolo+RCRalph+Redor144/test-cases-for-tram-stop-mapping
```

### Contributing code to main branch
In order to contribute code to the repository's main branch, create a pull request using your newly created branch and assign some reviewers to your code. When the pull request gets an approval from other members of this repository and all CI checks pass (if applicable), you will only then be able to merge it to main.

### Go environment setup
This repository uses Go 1.24.2. You can install Go by following the installation guide: https://go.dev/doc/install. On Linux, the recommended way of installing Go for development purposes is via Snap package due to ease of updates.

### Node.js environment setup
This repository uses Node.js v22.14.0. You can install Node.js by following the installation guide: https://nodejs.org/en/download. On Linux, the recommended way of installing Node.js is through Node Version Manager: https://github.com/nvm-sh/nvm.

### Pre-commit
Before committing to this repository, the developers should make sure that their code passes all required quality checks. In order to run them automatically, run:
```sh
pre-commit install
```

This command assumes `pre-commit` is available through the currently active Python virtual environment.

### Running GitHub Actions pipelines locally
In order to run pipelines locally, install [act](https://github.com/nektos/act), preferably using GitHub CLI:
```
gh extension install https://github.com/nektos/gh-act
```

You can then run the pipelines using:
```
gh act -P self-hosted=catthehacker/ubuntu:act-22.04
```

You can find more details about running pipelines locally in [act user guide](https://nektosact.com/).

### Running Wails in development mode
In order to run Wails in development mode, you need to install Wails CLI with:
``` bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

After successful installation, you should be able to successfully display the Wails CLI version:
```
wails version
```

If this command doesn't work, make sure to include the `go/bin` directory in your `PATH` environment variable.

In order to start the application in development mode, you should run:
```
wails dev
```

It might be necessary to use `webkit2_41` tag if the above command fails.

### Building binary using Wails
In order to build the binary, you need to specify the below `ldflags`:
| Value | Description | Example value | Default value |
| ----- | ----------- | ------------- | ------------- |
| `<module>/pkg/city.ServerURL` | Server URL base | https://tns-ee.rcralph.me | http://localhost:8000 |

Below you can find an example command which uses `ldflags`:
```bash
wails build -ldflags "-X github.com/TNSEngineerEdition/WailsClient/pkg/city.ServerURL=https://tns-ee.rcralph.me"
```

As above, it might be necessary to use `webkit2_41` tag if the above command fails.
