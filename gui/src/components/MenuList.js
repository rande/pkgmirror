// Copyright © 2016-present Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

import React from 'react'
import Avatar from 'material-ui/Avatar'
import { List, ListItem } from 'material-ui/List'
import Dashboard from 'material-ui/svg-icons/action/dashboard'
import Info from 'material-ui/svg-icons/action/info'

const MenuList = props => {
    const mirrorsItems = props.mirrors.map((mirror, pos) => (
        <ListItem
            key={pos}
            primaryText={mirror.SourceUrl}
            leftAvatar={
                <Avatar src={mirror.Icon} backgroundColor="rgba(0, 0, 0, 0)" />
            }
            onTouchTap={() => {
                props.onTouchStart(mirror)
            }}
            insetChildren={false}
        />
    ))

    const items = [
        <ListItem
            key="dashboard"
            primaryText="Dashboard"
            leftIcon={
                <Dashboard
                    viewBox="0 5 24 24"
                    style={{ width: 40, height: 40 }}
                />
            }
            onTouchTap={() => {
                props.homepage()
            }}
        />,

        ...mirrorsItems,

        <ListItem
            key="about"
            primaryText="About"
            leftIcon={
                <Info viewBox="0 4 24 24" style={{ width: 30, height: 30 }} />
            }
            onTouchTap={() => {
                props.about()
            }}
        />,
    ]

    return <List>{items}</List>
}

MenuList.propTypes = {
    mirrors: React.PropTypes.array,
    onTouchStart: React.PropTypes.func,
    homepage: React.PropTypes.func,
    about: React.PropTypes.func,
}

export default MenuList
