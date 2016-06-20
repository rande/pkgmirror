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
import { MirrorList, MenuList, CardMirror } from './redux/containers';

import { toggleDrawer, hideDrawer } from './redux/apps/guiApp';

import { Router, Route, IndexRoute } from 'react-router';
import { push } from 'react-router-redux';

const Container = props => (
    <div>{props.children}</div>
);

Container.propTypes = {
    children: React.PropTypes.any,
};

const Main = props => (<MuiThemeProvider muiTheme={props.Theme}>
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

            <MenuItem onTouchTap={props.homepage}>Mirrors</MenuItem>
            <MenuList />
        </Drawer>

        <Router history={props.history}>
            <Route path="/" component={Container}>
                <IndexRoute component={MirrorList} />
                <Route path="mirror/:id" component={CardMirror} />
            </Route>
        </Router>
    </div>
</MuiThemeProvider>);

Main.propTypes = {
    Theme: React.PropTypes.object,
    Title: React.PropTypes.string,
    DrawerOpen: React.PropTypes.bool,
    toggleDrawer: React.PropTypes.func,
    history: React.PropTypes.object,
    homepage: React.PropTypes.func,
};

const mapStateToProps = (state) => ({ ...state.guiApp });

const mapDispatchToProps = (dispatch) => ({
    toggleDrawer: () => {
        dispatch(toggleDrawer());
    },
    homepage: () => {
        dispatch(push('/'));
        dispatch(hideDrawer());
    },
});

export default connect(mapStateToProps, mapDispatchToProps)(Main);
