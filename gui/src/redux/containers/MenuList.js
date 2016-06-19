// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

import { connect } from 'react-redux';

import MenuList from '../../components/MenuList';

const mapStateToProps = (state) => ({
    mirrors: state.mirrorApp.mirrors,
});

export default connect(mapStateToProps)(MenuList);
