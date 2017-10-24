NPM mirroring
=============

Mirroring Workflow
------------------

> The url `https://registry.npmjs.org/-/all` is not available anymore, so it not possible to have
> a full copy of the npm's registry anymore.
> Pkgmirror will store on-demand and sync local packages.

1. On demand, the proxy will load the remote version if the package is new.
2. Update the local data with the modified date, if changed then update related package reference.
3. The package update will download the package information from ``https://registry.npmjs.org/package_name`` and update tarbal reference to point to the local entry point.


Entry Points
------------

* Get package information: ``/npm/package_name``
* Download archive: ``/npm/package_name/-/package_name-version.tgz``
