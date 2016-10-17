// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

import getMuiTheme from 'material-ui/styles/getMuiTheme';
import { deepOrange500 } from 'material-ui/styles/colors';
import { SMALL, MEDIUM, LARGE } from 'material-ui/utils/withWidth';

export const GUI_TOGGLE_DRAWER = 'GUI_TOGGLE_DRAWER';
export const GUI_HIDE_DRAWER = 'GUI_HIDE_DRAWER';
export const GUI_RESIZE_WINDOW = 'GUI_RESIZE_WINDOW';

export const toggleDrawer = () => ({
    type: GUI_TOGGLE_DRAWER,
});

export const hideDrawer = () => ({
    type: GUI_HIDE_DRAWER,
});

export const resizeApp = innerWidth => ({
    type: GUI_RESIZE_WINDOW, innerWidth,
});

const defaultState = {
    DrawerOpen: false,
    Title: 'PkgMirror',
    Theme: getMuiTheme({
        palette: {
            accent1Color: deepOrange500,
        },
    }),
    width: SMALL,
};

export function guiApp(state = defaultState, action) {
    switch (action.type) {
    case GUI_TOGGLE_DRAWER:
        return { ...state, DrawerOpen: !state.DrawerOpen };

    case GUI_HIDE_DRAWER:
        return { ...state, DrawerOpen: false };

    case GUI_RESIZE_WINDOW:
        const largeWidth = 992;
        const mediumWidth = 768;

        let width;

        if (action.innerWidth >= largeWidth) {
            width = LARGE;
        } else if (action.innerWidth >= mediumWidth) {
            width = MEDIUM;
        } else { // innerWidth < 768
            width = SMALL;
        }

        return { ...state, width };
    default:
        return state;
    }
}
