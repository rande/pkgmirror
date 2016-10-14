// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

import { connect } from 'react-redux';
import { push } from 'react-router-redux';

import MenuList from '../../components/MenuList';
import { hideDrawer } from '../apps/guiApp';

const mapStateToProps = state => ({
    mirrors: state.mirrorApp.mirrors,
});

const mapDispatchToProps = dispatch => ({
    homepage: () => {
        dispatch(push('/'));
        dispatch(hideDrawer());
    },
    about: () => {
        dispatch(push('/about'));
        dispatch(hideDrawer());
    },
    onTouchStart: (mirror) => {
        dispatch(push(`/mirror/${mirror.Id}`));
        dispatch(hideDrawer());
    },
});

export default connect(mapStateToProps, mapDispatchToProps)(MenuList);
