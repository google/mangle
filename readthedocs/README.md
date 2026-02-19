# Mangle documentation

[![Documentation Status](https://readthedocs.org/projects/mangle/badge/?version=latest)](https://readthedocs.org/projects/mangle/builds/)

This directory contains Mangle documentation sources.
This page describes how to generate to documentation.

If you want to read
the rendered documentation, go to [mangle.readthedocs.io](http://mangle.readthedocs.io).
## Overview

*Read the Docs* is a platform for hosting documentation. They offer a free
community plan to open source projects. Publishing is automated
through integration with Codeberg.

Mangle documentation uses *Read the Docs* with the following choices:

* [sphinx](https://www.sphinx-doc.org/en/master/) documentation generator
* [MyST markdown format](https://www.sphinx-doc.org/en/master/usage/markdown.html#markdown)

## Set up a virtual environment

A [virtual environment](https://docs.python.org/3/library/venv.html) helps
contain all python packages without having to deal with system-wide
installation.

```
> python -m venv manglereadthedocs
> . manglereadthedocs/bin/activate
(manglereadthedocs) > pip install -U sphinx
(manglereadthedocs) > READTHEDOCS=<path to readthedocs dir>
(manglereadthedocs) > pip install -r ${READTHEDOCS}/requirements.txt
```

## Building the documentation

```
> . manglereadthedocs/bin/activate
(manglereadthedocs) > READTHEDOCS=<path to readthedocs dir>
(manglereadthedocs) > sphinx-build -M html ${READTHEDOCS} output
```

The output can be found in the `output/html` directory.

You can then run a local http server to view the results:

```
(manglereadthedocs) > python3 -m http.server 8080
```
