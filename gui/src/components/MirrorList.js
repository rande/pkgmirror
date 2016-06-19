// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

import React from 'react';
import { List, ListItem } from 'material-ui/List';
import Avatar from 'material-ui/Avatar';
import ActionInfo from 'material-ui/svg-icons/action/info';
import ContentBlock from 'material-ui/svg-icons/content/block';
import PlayCircle from 'material-ui/svg-icons/av/loop';
import PauseCircle from 'material-ui/svg-icons/av/pause-circle-filled';

const MirrorList = props => {
    const { mirrors, events } = props;

    return (
        <List>{mirrors.map((mirror, pos) => {
            let rightIcon = <ActionInfo />;

            if (!mirror.Enabled) {
                rightIcon = <ContentBlock />;
            }

            let text = mirror.Type;

            if (mirror.Id in events) {
                if (events[mirror.Id].Status === 1) {
                    rightIcon = <PlayCircle />;
                } else {
                    if (events[mirror.Id].Status === 2) {
                        rightIcon = <PauseCircle />;
                    }
                }

                text += `' - ${events[mirror.Id].Message}`;
            }

            return (<ListItem
                key={pos}
                primaryText={mirror.SourceUrl}
                secondaryText={text}
                leftAvatar={<Avatar src={mirror.Icon} />}
                rightIcon={rightIcon}
            />);
        })}</List>
    );
};

MirrorList.propTypes = {
    mirrors: React.PropTypes.array,
    events: React.PropTypes.object,
};

export default MirrorList;
