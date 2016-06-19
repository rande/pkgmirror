// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

import React, {Component} from 'react';
import {Card, CardActions, CardHeader, CardMedia, CardTitle, CardText} from 'material-ui/Card';
import {List, ListItem} from 'material-ui/List';
import Avatar from 'material-ui/Avatar';
import ActionInfo from 'material-ui/svg-icons/action/info';
import ContentBlock from 'material-ui/svg-icons/content/block';
import PlayCircle from 'material-ui/svg-icons/av/loop';
import PauseCircle from 'material-ui/svg-icons/av/pause-circle-filled';

class MirrorList extends Component {
    render() {
        const {mirrors, events} = this.props;

        return (
            <List>{mirrors.map((mirror, pos) => {
                var rightIcon = <ActionInfo />;

                if (!mirror.Enabled) {
                    rightIcon = <ContentBlock />;
                }

                var text = mirror.Type;

                if (mirror.Id in events) {
                    if (events[mirror.Id].Status == 1) {
                        rightIcon = <PlayCircle />;
                    } else {
                        if (events[mirror.Id].Status == 2) {
                            rightIcon = <PauseCircle />;
                        }
                    }

                    text += " - " + events[mirror.Id].Message;
                }

                return <ListItem
                    key={pos}
                    primaryText={mirror.SourceUrl}
                    secondaryText={text}
                    leftAvatar={<Avatar src={mirror.Icon} />}
                    rightIcon={rightIcon}
                />
            })}</List>
        );
    }
}

export default MirrorList
