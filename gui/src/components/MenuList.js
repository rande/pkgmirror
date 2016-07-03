// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

import React from 'react';
import Avatar from 'material-ui/Avatar';
import { ListItem } from 'material-ui/List';

const MenuList = props => {
    const nestedItems = props.mirrors.map((mirror, pos) => (<ListItem
        key={pos}
        primaryText={mirror.SourceUrl}
        leftAvatar={<Avatar src={mirror.Icon} backgroundColor="rgba(0, 0, 0, 0)" />}
        onTouchTap={() => { props.onTouchStart(mirror); }}
    />));

    return (<ListItem
        primaryText="Mirrors"
        onTouchTap={() => { props.homepage(); }}
        initiallyOpen
        nestedItems={nestedItems}
    />);
};


MenuList.propTypes = {
    mirrors: React.PropTypes.array,
    onTouchStart: React.PropTypes.func,
    homepage: React.PropTypes.func,
};

export default MenuList;
