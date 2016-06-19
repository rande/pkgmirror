// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

import {connect} from 'react-redux';

import MirrorList from '../../components/MirrorList';

function mapStateToProps(state) {
    return {
        mirrors: state.mirrorApp.mirrors,
        events:  state.mirrorApp.events
    }
}

export default connect(mapStateToProps)(MirrorList)
