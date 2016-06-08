// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

import React, {Component} from 'react';
import {deepOrange500} from 'material-ui/styles/colors';
import getMuiTheme from 'material-ui/styles/getMuiTheme';
import MuiThemeProvider from 'material-ui/styles/MuiThemeProvider';
import AppBar from 'material-ui/AppBar';

import {MirrorList} from './components'

const styles = {
    container: {
        textAlign:  'center',
        paddingTop: 200,
    },
};

const muiTheme = getMuiTheme({
    palette: {
        accent1Color: deepOrange500,
    },
});

class Main extends Component {
    constructor(props, context) {
        super(props, context);
    }

    render() {

        return (
            <MuiThemeProvider muiTheme={muiTheme}>
                <div>
                    <AppBar
                        title="PkgMirror"
                        iconClassNameRight="muidocs-icon-navigation-expand-more"
                    />

                    <MirrorList />

                </div>

            </MuiThemeProvider>
        );
    }
}

export default Main;
