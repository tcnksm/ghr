## 0.5.4 (2017-01-11)

### Added 

- Allow to specify body [#55](https://github.com/tcnksm/ghr/pull/55)

## 0.5.3 (2016-10-31)

### Fixed

- Fix overlapping of DeleteRelease and CreateRelease #52

## 0.5.2 (2016-10-21)

### Fixed

- Output format

## 0.5.1 (2016-10-19)

### Fixed

- Can not use GitHub Enterprise environment #48
- Format verbs shows up on output #50

## 0.5.0 (2016-10)

### Added

- Nothing

### Deprecated

- `-stat` option

### Removed

- Nothing

### Fixed

- Refactoring & Adding a lots of tests

## 0.4.0 (2014-12-17)

### Added

- Support Github Enterprise (supported by [**@zoncoen**](https://github.com/zoncoen))
- Add more tests

### Deprecated

- Nothing

### Removed

- Nothing

### Fixed

- `--delete` sometimes breaks relase. This is temporary solution.

## 0.3.0 (2014-12-15)

### Added

- [goole/go-github](https://github.com/google/go-github) for GitHub API client
- `--stat` option to show how many tool downloaded
- Color output
- Many refactorings

### Deprecated

- Nothing

### Removed

- Old GitHub API client

### Fixed

- Nothing

## 0.2.0 (2014-12-09)

### Added

- Read `GITHUB_TOKEN` from `gitconfig` file ([**@sona-tar**](https://github.com/sona-tar), [#8](https://github.com/tcnksm/ghr/pull/8))
- When using `--delete` option, delete its git tag
- Support private repository ([**@virifi**](https://github.com/virifi), [#10](https://github.com/tcnksm/ghr/pull/10))
- Many refactoring

### Deprecated

- Nothing

### Removed

- Nothing

### Fixed

- Nothing

## 0.1.2 (2014-10-09)

### Added

- `--replace` option to replace artifact if it exist
- `--delete` option to delete release in advance if it exist
- [tcnksm/go-gitconfig](https://github.com/tcnksm/go-gitconfig) for extracting git config values

### Deprecated

- Nothing

### Removed

- Nothing

### Fixed

- Nothing

## 0.1.1 (2014-08-06)

### Added

- Limit amount of parallelism by the number of CPU
- `--username` option to set Github username
- `--token` option to set API token
- `--repository` option to set repository name
- `--draft` option to create unpublished release
- `--prerelease` option to create prelerease

### Deprecated

- Nothing

### Removed

- Nothing

### Fixed

- Nothing

## 0.1.0 (2014-07-29)

Initial release

### Added

- Add Fundamental features

### Deprecated

- Nothing

### Removed

- Nothing

### Fixed

- Nothing
