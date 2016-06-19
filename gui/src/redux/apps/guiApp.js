// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

import getMuiTheme from 'material-ui/styles/getMuiTheme';
import {deepOrange500} from 'material-ui/styles/colors';

export const GUI_TOGGLE_DRAWER = 'GUI_TOGGLE_DRAWER';

export const toggleDrawer = () => ({
    type: GUI_TOGGLE_DRAWER
});

const styles = {
    container: {
        textAlign:  'center',
        paddingTop: 200
    }
};

const defaultState = {
    DrawerOpen: false,
    Title:      "PkgMirror",
    Theme:      getMuiTheme({
        palette: {
            accent1Color: deepOrange500
        }
    })
};

export function guiApp(state = defaultState, action) {
    switch (action.type) {
        case GUI_TOGGLE_DRAWER:
            return {...state, DrawerOpen: !state.DrawerOpen};
    }

    return state;
}