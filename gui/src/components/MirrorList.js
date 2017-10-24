// Copyright © 2016-present Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

import React from 'react';
import { GridList, GridTile } from 'material-ui/GridList';
import { MEDIUM, LARGE } from 'material-ui/utils/withWidth';

import ActionInfo from 'material-ui/svg-icons/action/info';
import ContentBlock from 'material-ui/svg-icons/content/block';
import PlayCircle from 'material-ui/svg-icons/av/loop';
import PauseCircle from 'material-ui/svg-icons/av/pause-circle-filled';

const MirrorList = (props) => {
    const { mirrors, events } = props;

    let cols = 2,
        cellHeight = 150;

    if (props.width === MEDIUM) {
        cols = 2;
        cellHeight = 200;
    }

    if (props.width === LARGE) {
        cols = 4;
        cellHeight = 300;
    }

    return (
        <GridList
            cols={cols}
            cellHeight={cellHeight}
            padding={1}
        >{mirrors.map((mirror, pos) => {
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

                text += ` - ${events[mirror.Id].Message}`;
            }

            return (<GridTile
                key={pos}
                title={mirror.SourceUrl}
                subtitle={text}
                titlePosition="bottom"
                actionIcon={rightIcon}
                actionPosition="left"
                onClick={() => { props.onTouchStart(mirror); }}
            >
                <img src={mirror.Icon} role="presentation" />
            </GridTile>)
            ;
        })}</GridList>
    );
};

MirrorList.propTypes = {
    mirrors: React.PropTypes.array,
    events: React.PropTypes.object,
    onTouchStart: React.PropTypes.func,
    width: React.PropTypes.int,
};

export default MirrorList;
