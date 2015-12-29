# Gist-it

## Description

A small Go based tool to fetch and create Gists on Github.

## Dependencies

Install with `go get [DEPENDENCY]`:

- github.com/codegangsta/cli
- github.com/google/go-github/github
- golang.org/x/oauth2


# Building

- git clone github.com/hartfordfive/gist-it.git
- cd gist-it
- go build


# Usage

Create a gist with one ore many files (*The description and public status will be prompted for after*)
```Go
gist-it create [filename1] [[filename2] [[filename3]]]
```

List your all your current Gists
```Go
gist-it list
```

Get a specific Gist
```Go
gist-it get [GIST_ID]
```

Get help for commands
```Go
gist-it help
```


## Bugs & Feature Requests

Please open an issue for any bugs or feature requests.


## Author

Alain Lefebvre  (hartfordfive@gmail.com)


## License

Covered under the MIT License
