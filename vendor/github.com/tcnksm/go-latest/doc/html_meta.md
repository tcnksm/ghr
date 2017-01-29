# HTML meta tag version discovery

`go-latest.HTMLMeta` uses HTML meta tag to check latest version of your tool. It will request provided `URL` and inspec the HTML returned for meta tags that have the following format:

```bash
<meta name="go-latest" content="product-name SemVer message">
```

- `product-name` is your tool name. It MUST be filled
- `SemVer` is your tool version by [Semantic Versioning](http://semver.org/). It MUST be filled
- `message` is a message. It MAY be filled

For example, if you want to check latest version of `reduce-worker`, you just prepare a HTML page which contains following tags.

```bash
<meta name="go-latest" content="reduce-worker 1.2.3">
```

You can know latest version is `1.2.3`. 

## References

`go-latest`'s HTML meta tag version discovery specification refers following:

- [Golang Remote import paths](https://golang.org/cmd/go/#hdr-Remote_import_paths)
- [App Container Image Discovery](https://github.com/appc/spec/blob/master/SPEC.md#app-container-image-discovery)



