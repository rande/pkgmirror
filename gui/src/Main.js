// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

import React from 'react';
import MuiThemeProvider from 'material-ui/styles/MuiThemeProvider';
import AppBar from 'material-ui/AppBar';
import Drawer from 'material-ui/Drawer';
import {SMALL, MEDIUM, LARGE} from 'material-ui/utils/withWidth';
import spacing from 'material-ui/styles/spacing';

import { connect } from 'react-redux';
import { MirrorList, MenuList, CardMirror } from './redux/containers';
import { List } from 'material-ui/List';

import { toggleDrawer, hideDrawer } from './redux/apps/guiApp';

import { Router, Route, IndexRoute } from 'react-router';
import { push } from 'react-router-redux';

const Container = props => (
    <div>{props.children}</div>
);

Container.propTypes = {
    children: React.PropTypes.any,
};

const Main = props => {

    let DrawerOpen = false;
    let marginLeft = 0;
    if (props.width == SMALL && props.DrawerOpen) {
        DrawerOpen = true;
    }

    if (props.width != SMALL) {
        DrawerOpen = true;
        marginLeft = 300;
    }

    return <MuiThemeProvider muiTheme={props.Theme}>
        <div>
            <AppBar
                title={props.Title}
                iconClassNameRight="muidocs-icon-navigation-expand-more"
                onLeftIconButtonTouchTap={props.toggleDrawer}
                showMenuIconButton={props.width == SMALL}
            />

            <Drawer open={DrawerOpen} docked={true} width={300}>
                <AppBar
                    title={props.Title}
                    onLeftIconButtonTouchTap={props.toggleDrawer}
                    showMenuIconButton={props.width == SMALL}
                />
                <List>
                    <MenuList />
                </List>
            </Drawer>

            <div className="foobar" style={{marginLeft: marginLeft + 'px'}}>
                <Router history={props.history} >
                    <Route path="/" component={Container}>
                        <IndexRoute component={MirrorList} />
                        <Route path="mirror/:id" component={CardMirror} />
                    </Route>
                </Router>
            </div>
        </div>
    </MuiThemeProvider>
};

Main.propTypes = {
    Theme: React.PropTypes.object,
    Title: React.PropTypes.string,
    DrawerOpen: React.PropTypes.bool,
    toggleDrawer: React.PropTypes.func,
    history: React.PropTypes.object,
    homepage: React.PropTypes.func,
    width: React.PropTypes.number.isRequired,
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
