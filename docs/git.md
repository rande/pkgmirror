Git Mirroring
=============

The git mirroring service allows to mirror git repositories into a local server.

Mirroring Workflow
------------------

All repositories are stored in the ``DataDir/git/hostname`` path. So for github.com it will be: ``DataDir/git/github.com``.

1. Iterate over each hostname
2. Start a goroutinne for each hostname
3. Iterate over folder ending by ``.git`` (up to 3 nested levels)
4. Run the ``fetch`` and ``update-server-info`` commands on each mirror

Entry Points
------------

### Clone repository

The current implementation provides support for the [dump http protocol](https://git-scm.com/book/tr/v2/Git-on-the-Server-The-Protocols), so
 it is possible to only clone over http/https.
 
    git clone https://mirror.example.com/git/github.com/rande/pkgmirror.git
    
### Archive

You can also download a zip for a specific version:

    curl https://mirror.example.com/git/github.com/rande/pkgmirror/master.git
    curl https://mirror.example.com/git/github.com/rande/pkgmirror/9c34490d5fb421d45bb8634b84308995b407fb4b.git

Please note, only semver tags and commits are cached.

