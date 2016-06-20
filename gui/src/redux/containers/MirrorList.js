// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

import { connect } from 'react-redux';

import MirrorList from '../../components/MirrorList';
import { push } from 'react-router-redux';
import { hideDrawer } from '../apps/guiApp';

const mapStateToProps = (state) => ({
    mirrors: state.mirrorApp.mirrors,
    events: state.mirrorApp.events,
});

const mapDispatchToProps = (dispatch) => ({
    onTouchStart: (mirror) => {
        dispatch(push(`/mirror/${mirror.Id}`));
        dispatch(hideDrawer());
    },
});

export default connect(mapStateToProps, mapDispatchToProps)(MirrorList);
