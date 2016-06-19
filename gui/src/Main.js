// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

import React from 'react';
import MuiThemeProvider from 'material-ui/styles/MuiThemeProvider';
import AppBar from 'material-ui/AppBar';
import Drawer from 'material-ui/Drawer';
import MenuItem from 'material-ui/MenuItem';

import { connect } from 'react-redux';
import { MirrorList, MenuList } from './redux/containers';
import { toggleDrawer } from './redux/apps/guiApp';

const Main = props => (
    <MuiThemeProvider muiTheme={props.Theme}>
        <div>
            <AppBar
                title={props.Title}
                iconClassNameRight="muidocs-icon-navigation-expand-more"
                onLeftIconButtonTouchTap={props.toggleDrawer}
            />

            <Drawer open={props.DrawerOpen}>
                <AppBar
                    title={props.Title}
                    onLeftIconButtonTouchTap={props.toggleDrawer}
                />

                <MenuItem>Mirrors</MenuItem>
                <MenuList />
                <MenuItem>About</MenuItem>
            </Drawer>

            <MirrorList />
        </div>

    </MuiThemeProvider>
);

Main.propTypes = {
    Theme: React.PropTypes.object,
    Title: React.PropTypes.string,
    DrawerOpen: React.PropTypes.bool,
    toggleDrawer: React.PropTypes.func,
};

const mapStateToProps = (state) => ({ ...state.guiApp });

const mapDispatchToProps = (dispatch) => ({
    toggleDrawer: (id) => {
        dispatch(toggleDrawer(id));
    },
});

export default connect(mapStateToProps, mapDispatchToProps)(Main);
