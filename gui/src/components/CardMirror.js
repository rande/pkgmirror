// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

import React from 'react';

import { Card, CardHeader, CardText } from 'material-ui/Card';
import Markdown from 'react-markdown';
import Avatar from 'material-ui/Avatar';

const CardMirror = props => (
    <Card>
        <CardHeader
            title={props.mirror.SourceUrl}
            subtitle={props.mirror.Type}
            avatar={<Avatar src={props.mirror.Icon} backgroundColor="rgba(0, 0, 0, 0)" />}
        />
        <CardText>
            <Markdown source={props.mirror.Usage} />
        </CardText>
    </Card>
);

CardMirror.propTypes = {
    mirror: React.PropTypes.object,
};

export default CardMirror;
