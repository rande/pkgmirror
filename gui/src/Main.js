// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

import React, {Component} from 'react';
import MuiThemeProvider from 'material-ui/styles/MuiThemeProvider';
import AppBar from 'material-ui/AppBar';
import Drawer from 'material-ui/Drawer';
import MenuItem from 'material-ui/MenuItem';

import {connect} from 'react-redux';
import {MirrorList, MenuList} from './redux/containers';
import {toggleDrawer} from './redux/apps/guiApp';

class Main extends Component {
    render() {
        return (
            <MuiThemeProvider muiTheme={this.props.Theme}>
                <div>
                    <AppBar
                        title={this.props.Title}
                        iconClassNameRight="muidocs-icon-navigation-expand-more"
                        onLeftIconButtonTouchTap={this.props.toggleDrawer}
                    />

                    <Drawer open={this.props.DrawerOpen}>
                        <AppBar title={this.props.Title}
                                onLeftIconButtonTouchTap={this.props.toggleDrawer}
                        />

                        <MenuItem>Mirrors</MenuItem>
                        <MenuList />
                        <MenuItem>About</MenuItem>
                    </Drawer>

                    <MirrorList />
                </div>

            </MuiThemeProvider>
        );
    };
}

const mapStateToProps = (state) => ({...state.guiApp});

const mapDispatchToProps = (dispatch) => ({
    toggleDrawer: (id) => {
        dispatch(toggleDrawer(id))
    }
});

export default connect(mapStateToProps, mapDispatchToProps)(Main);
