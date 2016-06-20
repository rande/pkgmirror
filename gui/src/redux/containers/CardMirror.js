// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

import { connect } from 'react-redux';

import CardMirror from '../../components/CardMirror';

const mapStateToProps = (state, ownProps) => {
    let mirror = {};

    state.mirrorApp.mirrors.every((v) => {
        if (ownProps.params.id === v.Id) {
            mirror = v;

            return false;
        }

        return true;
    });

    return {
        mirror,
    };
};

export default connect(mapStateToProps)(CardMirror);
