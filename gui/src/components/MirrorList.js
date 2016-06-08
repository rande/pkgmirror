// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

import React, {Component} from 'react';
import {Card, CardActions, CardHeader, CardMedia, CardTitle, CardText} from 'material-ui/Card';
import {List, ListItem} from 'material-ui/List';
import Avatar from 'material-ui/Avatar';
import ActionInfo from 'material-ui/svg-icons/action/info';
import Transmit from "react-transmit";

class MirrorList extends Component {
    constructor(props, context) {
        super(props, context);
    }

    render() {
        const {mirrors} = this.props;

        return (
            <List>{mirrors.map((mirror, pos) => <ListItem
                    key={pos}
                    primaryText={mirror.SourceUrl}
                    secondaryText={mirror.Type}
                    leftAvatar={<Avatar src={mirror.Icon} />}
                    rightIcon={<ActionInfo />}
                />
            )}</List>
        );
    }
}

export default Transmit.createContainer(MirrorList, {
    initialVariables: {},
    fragments:        {
        mirrors () {
            return fetch("/api/mirrors").then(res => res.json());
        }
    }
});
