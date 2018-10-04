// Copyright © 2016-present Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

import React from 'react'
import { render } from 'react-dom'
import { createStore, combineReducers, applyMiddleware } from 'redux'
import { Provider } from 'react-redux'
import {
    syncHistoryWithStore,
    routerReducer,
    routerMiddleware,
} from 'react-router-redux'

import injectTapEventPlugin from 'react-tap-event-plugin'

import { hashHistory } from 'react-router'

import Main from './Main' // Our custom react component
import { mirrorApp, addList, updateState } from './redux/apps/mirrorApp'
import { guiApp, resizeApp } from './redux/apps/guiApp'

// Needed for onTouchTap
// http://stackoverflow.com/a/34015469/988941
injectTapEventPlugin()

const middleware = routerMiddleware(hashHistory)
const reducers = combineReducers({
    mirrorApp,
    guiApp,
    routing: routerReducer,
})

const store = createStore(reducers, applyMiddleware(middleware))

if (window) {
    let deferTimer
    window.addEventListener('resize', () => {
        clearTimeout(deferTimer)
        deferTimer = setTimeout(() => {
            store.dispatch(resizeApp(window.innerWidth))
        }, 200)
    })

    // init the state
    store.dispatch(resizeApp(window.innerWidth))
}

const history = syncHistoryWithStore(hashHistory, store)
// history.listen(location => {
//     // console.log("Render Main", location);
// });

fetch('/api/mirrors').then(res => {
    res.json().then(data => {
        store.dispatch(addList(data))
    })
})

const ev = new EventSource('/api/sse')
ev.onmessage = em => {
    try {
        const data = JSON.parse(em.data)
        store.dispatch(updateState(data))
    } catch (e) {
        console.error(e)
    }
}

render(
    <Provider store={store}>
        <Main history={history} />
    </Provider>,
    document.getElementById('app')
)
