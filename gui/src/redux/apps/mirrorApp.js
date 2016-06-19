// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

export const MIRROR_ADD_LIST = 'MIRROR_ADD_LIST';
export const MIRROR_UPDATE_STATE = 'MIRROR_UPDATE_STATE';

export const updateState = mirror => ({
    type: MIRROR_UPDATE_STATE, mirror,
});

export const addList = list => ({
    type: MIRROR_ADD_LIST, list,
});

const defaultState = {
    mirrors: [{
        Id: 'fake.id',
        Type: 'redux',
        Name: 'github',
        SourceUrl: 'http://redux.js.org',
        TargetUrl: 'http://redux.js.org',
        Icon: 'http://freeiconbox.com/icon/256/34429.png',
        Enabled: true,
    }],
    events: {},
};

export function mirrorApp(state = defaultState, action) {
    switch (action.type) {
    case MIRROR_ADD_LIST:
        return { ...state, mirrors: action.list };

    case MIRROR_UPDATE_STATE: {
        const s = { ...state };
        s.events = { ...state.events };
        s.events[action.mirror.Id] = action.mirror;

        return s;
    }

    default:
        return state;
    }
}
