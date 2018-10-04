// Copyright © 2016-present Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

import React from 'react'

const About = () => (
    <div style={{ padding: 10 }}>
        <h1>PkgMirror</h1>
        <p>
            This project has been created by{' '}
            <a href="https://thomas.rabaix.net">Thomas Rabaix</a> to avoid
            downtime while working with remote dependencies. <br />
            <br />
            The backend is coded using the Golang programming language for
            syncing the different sources. On the frontend, ReactJS is used with
            Material-UI. <br />
            <br />
            Feel free to contribute on:{' '}
            <a href="https://github.com/rande/pkgmirror">
                https://github.com/rande/pkgmirror
            </a>
            .
        </p>
        <h2>Licence</h2>
        Copyright (c) 2016 Thomas Rabaix <br />
        <br />
        Permission is hereby granted, free of charge, to any person obtaining a
        copy of this software and associated documentation files (the
        "Software"), to deal in the Software without restriction, including
        without limitation the rights to use, copy, modify, merge, publish,
        distribute, sublicense, and/or sell copies of the Software, and to
        permit persons to whom the Software is furnished to do so, subject to
        the following conditions: <br />
        <br />
        The above copyright notice and this permission notice shall be included
        in all copies or substantial portions of the Software.
        <br />
        <br />
        THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS
        OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
        MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
        IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
        CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
        TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
        SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
    </div>
)

export default About
