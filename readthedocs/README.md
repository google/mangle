# Mangle documentation

[![Documentation Status](https://readthedocs.org/projects/mangle/badge/?version=latest)](https://readthedocs.org/projects/mangle/builds/)

This repository contains Mangle documentation sources. Rendered documentation is hosted on [mangle.readthedocs.io](http://mangle.readthedocs.io).

## Setting up an environment

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
