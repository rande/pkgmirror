// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

import React from 'react';
import Avatar from 'material-ui/Avatar';
import MenuItem from 'material-ui/MenuItem';
import Menu from 'material-ui/Menu';

const MenuList = props => (
    <Menu autoWidth>
        {props.mirrors.map((mirror, pos) => (<MenuItem
            key={pos}
            primaryText={mirror.SourceUrl}
            leftIcon={<Avatar src={mirror.Icon} />}
            onTouchTap={() => { props.onTouchStart(mirror); }}
        />))}
    </Menu>
);

MenuList.propTypes = {
    mirrors: React.PropTypes.array,
    onTouchStart: React.PropTypes.func,
};

export default MenuList;
