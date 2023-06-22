# lurch

Simple CI/CD

![](screenshot.png)

### Usage
```
Usage: lurch [options]
Options:
	-h, --help			print this help
	-v, --version			print version
	-t, --path [PATH]		absolute path to work dir
	-p, --port [PORT]		sets port for listening
	-a, --app-url [APP_URL]		application url (if behind proxy)
```

## How to setup project
1. In `workdir` create a folder with name that represents the project.
2. In created folder create `script.sh` and add execute permission to it. Or for Windows create `script.cmd`.
3. Create your shell script with some content e.g.
```bash
#!/bin/sh -e

git clone repository ./

make test

make clean build

scp target/build server:/opt/www

rm -rf .git/
```
4. Open lurch in browser and start the job.