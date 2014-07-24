ghr
====

Easy to release your project to Github Release page.

## Synopsis

`ghr` enable you to create release on Github and upload your artifacts to it. 

## Demo

## Usage

To upload all package in `pkg` directory:

```bash
$ ghr v1.0 pkg
```

## Install

To install, use `go get`:

```bash
$ go get -d github.com/tcnksm/ghr
```

## VS.

- [aktau/github-release](https://github.com/aktau/github-release) - `github-release` can also create and edit releases and upload artifacts. It has many options. `ghr` is a super slim alternative.

## Contribution

1. Fork ([https://github.com/tcnksm/ghr/fork](https://github.com/tcnksm/ghr/fork))
1. Create a feature branch
1. Commit your changes
1. Rebase your local changes against the master branch
1. Run test suite with the `go test ./...` command and confirm that it passes
1. Run `gofmt -s`
1. Create new Pull Request

## Licence

[MIT](https://github.com/tcnksm/ghr/blob/master/LICENCE)

## Author

[tcnksm](https://github.com/tcnksm)
