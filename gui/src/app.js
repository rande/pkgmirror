// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

import React from 'react';
import { render } from 'react-dom';
import { createStore, combineReducers } from 'redux';
import { Provider } from 'react-redux';

import injectTapEventPlugin from 'react-tap-event-plugin';
import Main from './Main'; // Our custom react component

import { mirrorApp, addList, updateState } from './redux/apps/mirrorApp';
import { guiApp } from './redux/apps/guiApp';

// Needed for onTouchTap
// http://stackoverflow.com/a/34015469/988941
injectTapEventPlugin();

let store = createStore(combineReducers({ mirrorApp, guiApp }));

fetch('/api/mirrors').then(res => {
    res.json().then(data => {
        store.dispatch(addList(data));
    });
});

const ev = new EventSource('/api/sse');
ev.onmessage = (em) => {
    try {
        const data = JSON.parse(em.data);
        store.dispatch(updateState(data));
    } catch (e) {
        console.error(e);
    }
};

render(<Provider store={store}><Main /></Provider>, document.getElementById('app'));
