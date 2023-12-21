# lurch

Simple CI/CD without database or any dependency

![](screenshot.png)

## Usage
```
Usage: lurch [options]
Options:
	-h, --help			print this help
	-v, --version			print version
	-t, --path [PATH]		absolute path to work dir
	-p, --port [PORT]		sets port for listening
	-a, --app-url [APP_URL]		application url (if behind proxy)
	-n, --name [NAME]		name of application to be displayed
	-sj, --start-job [PROJECT]	makes client call to origin server and starts the build of [PROJECT]
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

### Start build from script of different project
Inside your project build script call lurch with parameter `-sj` followed by project name. If build is started, the result code of `lurch -sj` is `0`, otherwise it is `1`.

```bash
#!/bin/sh -e

git clone repository ./

make clean build

/usr/bin/lurch -sj repository-deploy

```

## Roadmap
- [x] Core (0.1.0)
- [x] REST API (0.1.0)
- [x] Web UI (0.1.0)
- [X] Build parameters passed as environmentals (0.2.0)
- [X] Hide dot project (e.g. `.ignored-project`) (0.3.0)
- [X] Dark theme (0.3.0)
- [X] Custom name of application (0.3.0)
- [ ] Periodical watcher (running custom script saving state of last check)
- [ ] Pipelining (jobs started according the result status)
- [X] Starting build from build script
- [X] Size and existance of artifact