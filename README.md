# devn

Lightweight tools which combine to create a continuous delivery system.

The working parts at this stage:

## devn-run

An extension to bash scripts using comments, so it is backwards compatible.

A line like

```
#+node flag1
```

will run the given command in the environments/ folder, i.e.

```
environments/node flag1
```

Then pipe the rest of the script to that process until the next #+ line or EOF
The idea is that each environment file will boot a bash environment, either through ssh 
or in a docker container or whatever, in which the rest of the commands should be run.

The section prior to the first line is run as an ordinary shell


Example script
```
#!/bin/bash

git clone ssh://git@github.com/private/dependency to/somewhere

#+node
npm install
webpack

#+golang
go build -o server to/somewhere/server.go
```
Which will run just fine on your dev machine which has go and node installed, but when run through the parser on a build server, will call the `node` and `golang` scripts, which could look like:

```
#!/bin/bash

docker run --rm -i\
	--volume `pwd`:/app \
	--workdir /app \
	node:5.11 /bin/bash $1
```

Environment variables from the parent shell are passed down, and no attempt has been made for security.


## devn-docker

An extension to the Dockerfile, also through comments, which is used to run scripts in the docker-ext folder.

Actually it knows nothing about docker, it's really just parsing comments, so could work on any file which has comments.

Lines like

```
#+GIT ssh://git@github.com/awesome/project
```

will call

```
docker-ext/GIT ssh://git@github.com/awesome/project
```

(which could just be a symlink to git, but you probably want to whitelist things, or create bash scripts, or even use the devn parser above)

# Notes

## Multi-line

Both parsers support multi-line strings by escaping the newline, however, since they should remain valid bash scripts and Docker files, the next line should also be a comment, and the comment symbol will be dropped, e.g.

```
#+GIT \
# ssh://git@github.com/awesome/project
```
Resolves to
```
docker-ext/GIT  ssh://git@github.com/awesome/project
```

It's not required, if the # is missing, it will still merge the lines, but that's probably not the behavior you are looking for


### Security

There is no attempt made for adding security, your PATH variable, running arbitrary commands etc, this is designed for somewhat trusted scripts.
