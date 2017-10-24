// Copyright © 2016-present Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

import { connect } from 'react-redux';
import { push } from 'react-router-redux';

import MirrorList from '../../components/MirrorList';
import { hideDrawer } from '../apps/guiApp';

const mapStateToProps = state => ({
    mirrors: state.mirrorApp.mirrors,
    events: state.mirrorApp.events,
    width: state.guiApp.width,
});

const mapDispatchToProps = dispatch => ({
    onTouchStart: (mirror) => {
        dispatch(push(`/mirror/${mirror.Id}`));
        dispatch(hideDrawer());
    },
    homepage: () => {
        dispatch(push('/'));
    },
});

export default connect(mapStateToProps, mapDispatchToProps)(MirrorList);
