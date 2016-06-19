// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

import React, {Component} from 'react';
import {Card, CardActions, CardHeader, CardMedia, CardTitle, CardText} from 'material-ui/Card';
import {List, ListItem} from 'material-ui/List';
import Avatar from 'material-ui/Avatar';
import MenuItem from 'material-ui/MenuItem';
import Menu from 'material-ui/Menu';

class MenuList extends Component {
    render() {
        const {mirrors} = this.props;

        return (
            <Menu autoWidth={true}>{mirrors.map((mirror, pos) => {
                return <MenuItem
                    key={pos}
                    primaryText={mirror.SourceUrl}
                    leftIcon={<Avatar src={mirror.Icon} />}
                />
            })}</Menu>
        );
    }
}

export default MenuList