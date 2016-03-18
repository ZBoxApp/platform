// Copyright (c) 2015 Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

import $ from 'jquery';
import ReactDOM from 'react-dom';
import NewChannelFlow from './new_channel_flow.jsx';
import MoreDirectChannels from './more_direct_channels.jsx';
import SidebarHeader from './sidebar_header.jsx';
import UnreadChannelIndicator from './unread_channel_indicator.jsx';
import TutorialTip from './tutorial/tutorial_tip.jsx';

import ChannelStore from 'stores/channel_store.jsx';
import UserStore from 'stores/user_store.jsx';
import TeamStore from 'stores/team_store.jsx';
import PreferenceStore from 'stores/preference_store.jsx';

import * as AsyncClient from 'utils/async_client.jsx';
import * as Utils from 'utils/utils.jsx';

import Constants from 'utils/constants.jsx';

import {FormattedMessage, FormattedHTMLMessage} from 'react-intl';

const Preferences = Constants.Preferences;
const TutorialSteps = Constants.TutorialSteps;

import {Tooltip, OverlayTrigger} from 'react-bootstrap';
import loadingGif from 'images/load.gif';

import React from 'react';
import {browserHistory, Link} from 'react-router';

import favicon from 'images/favicon/favicon-16x16.png';
import redFavicon from 'images/favicon/redfavicon-16x16.png';

export default class Sidebar extends React.Component {
    constructor(props) {
        super(props);

        this.badgesActive = false;
        this.firstUnreadChannel = null;
        this.lastUnreadChannel = null;

        this.getStateFromStores = this.getStateFromStores.bind(this);

        this.onChange = this.onChange.bind(this);
        this.onScroll = this.onScroll.bind(this);
        this.updateUnreadIndicators = this.updateUnreadIndicators.bind(this);
        this.handleLeaveDirectChannel = this.handleLeaveDirectChannel.bind(this);
        this.handleResize = this.handleResize.bind(this);

        this.showMoreChannelsModal = this.showMoreChannelsModal.bind(this);
        this.showNewChannelModal = this.showNewChannelModal.bind(this);
        this.hideNewChannelModal = this.hideNewChannelModal.bind(this);
        this.showMoreDirectChannelsModal = this.showMoreDirectChannelsModal.bind(this);
        this.hideMoreDirectChannelsModal = this.hideMoreDirectChannelsModal.bind(this);

        this.createChannelElement = this.createChannelElement.bind(this);
        this.updateTitle = this.updateTitle.bind(this);

        this.isLeaving = new Map();

        const state = this.getStateFromStores();
        state.newChannelModalType = '';
        state.showDirectChannelsModal = false;
        state.loadingDMChannel = -1;
        state.windowWidth = Utils.windowWidth();
        this.state = state;
    }
    getTotalUnreadCount() {
        let msgs = 0;
        let mentions = 0;
        const unreadCounts = this.state.unreadCounts;

        Object.keys(unreadCounts).forEach((chId) => {
            msgs += unreadCounts[chId].msgs;
            mentions += unreadCounts[chId].mentions;
        });

        return {msgs, mentions};
    }
    getStateFromStores() {
        const members = ChannelStore.getAllMembers();
        const currentChannelId = ChannelStore.getCurrentId();
        const currentUserId = UserStore.getCurrentId();
        const currentUser = UserStore.getCurrentUser();
        const isGuest = currentUser ? Utils.isGuest(currentUser.roles) : false;

        const channels = Object.assign([], ChannelStore.getAll());
        channels.sort((a, b) => a.display_name.localeCompare(b.display_name));

        const publicChannels = channels.filter((channel) => channel.type === Constants.OPEN_CHANNEL);
        const privateChannels = channels.filter((channel) => channel.type === Constants.PRIVATE_CHANNEL);

        const preferences = PreferenceStore.getCategory(Constants.Preferences.CATEGORY_DIRECT_CHANNEL_SHOW);

        const directChannels = [];
        for (const [name, value] of preferences) {
            if (value !== 'true') {
                continue;
            }

            const teammateId = name;

            let directChannel = channels.find(Utils.isDirectChannelForUser.bind(null, teammateId));

            // a direct channel doesn't exist yet so create a fake one
            if (directChannel == null) {
                directChannel = {
                    name: Utils.getDirectChannelName(currentUserId, teammateId),
                    last_post_at: 0,
                    total_msg_count: 0,
                    type: Constants.DM_CHANNEL,
                    fake: true
                };
            } else {
                directChannel = JSON.parse(JSON.stringify(directChannel));
            }

            directChannel.display_name = Utils.displayUsername(teammateId);
            directChannel.teammate_id = teammateId;
            directChannel.status = UserStore.getStatus(teammateId);

            directChannels.push(directChannel);
        }

        directChannels.sort(this.sortChannelsByDisplayName);

        const hiddenDirectChannelCount = UserStore.getActiveOnlyProfileList(true).
            filter((user) => {
                return !Utils.isGuest(user.roles);
            }).length - directChannels.length;

        const tutorialStep = PreferenceStore.getInt(Preferences.TUTORIAL_STEP, UserStore.getCurrentId(), 999);

        let channelMembers;
        if (isGuest) {
            channelMembers = ChannelStore.getExtraInfo(currentChannelId).members.filter((m) => {
                return m.id !== currentUserId;
            });
        }

        return {
            activeId: currentChannelId,
            members,
            publicChannels,
            privateChannels,
            directChannels,
            hiddenDirectChannelCount,
            unreadCounts: JSON.parse(JSON.stringify(ChannelStore.getUnreadCounts())),
            showTutorialTip: tutorialStep === TutorialSteps.CHANNEL_POPOVER,
            currentTeam: TeamStore.getCurrent(),
            currentUser: currentUser,
            channelMembers,
            isGuest
        };
    }

    componentDidMount() {
        ChannelStore.addChangeListener(this.onChange);
        UserStore.addChangeListener(this.onChange);
        UserStore.addStatusesChangeListener(this.onChange);
        TeamStore.addChangeListener(this.onChange);
        PreferenceStore.addChangeListener(this.onChange);

        this.updateTitle();
        this.updateUnreadIndicators();

        window.addEventListener('resize', this.handleResize);

        if (this.state.isGuest) {
            const membersIds = this.state.channelMembers.map((m) => {
                return m.id;
            });
            this.intervalId = setInterval(() => AsyncClient.getMemberStatuses(membersIds), 30000);
        }
    }
    shouldComponentUpdate(nextProps, nextState) {
        if (!Utils.areObjectsEqual(nextState, this.state)) {
            return true;
        }
        return false;
    }
    componentDidUpdate() {
        this.updateTitle();
        this.updateUnreadIndicators();
        if (!Utils.isMobile()) {
            $('.sidebar--left .nav-pills__container').perfectScrollbar();
        }
    }
    componentWillUnmount() {
        window.removeEventListener('resize', this.handleResize);

        ChannelStore.removeChangeListener(this.onChange);
        UserStore.removeChangeListener(this.onChange);
        UserStore.removeStatusesChangeListener(this.onChange);
        TeamStore.removeChangeListener(this.onChange);
        PreferenceStore.removeChangeListener(this.onChange);

        if (this.intervalId) {
            clearInterval(this.intervalId);
        }
    }
    handleResize() {
        this.setState({
            windowWidth: Utils.windowWidth(),
            windowHeight: Utils.windowHeight()
        });
    }
    onChange() {
        this.setState(this.getStateFromStores());
    }
    updateTitle() {
        const channel = ChannelStore.getCurrent();
        if (channel && this.state.currentTeam) {
            let currentSiteName = '';
            if (global.window.mm_config.SiteName != null) {
                currentSiteName = global.window.mm_config.SiteName;
            }

            let currentChannelName = channel.display_name;
            if (channel.type === 'D') {
                const teammate = Utils.getDirectTeammate(channel.id);
                if (teammate != null) {
                    currentChannelName = teammate.username;
                }
            }

            const unread = this.getTotalUnreadCount();
            const mentionTitle = unread.mentions > 0 ? '(' + unread.mentions + ') ' : '';
            const unreadTitle = unread.msgs > 0 ? '* ' : '';
            document.title = mentionTitle + unreadTitle + currentChannelName + ' - ' + this.state.currentTeam.display_name + ' ' + currentSiteName;
        }
    }
    onScroll() {
        this.updateUnreadIndicators();
    }
    updateUnreadIndicators() {
        const container = $(ReactDOM.findDOMNode(this.refs.container));

        var showTopUnread = false;
        var showBottomUnread = false;

        if (this.firstUnreadChannel) {
            var firstUnreadElement = $(ReactDOM.findDOMNode(this.refs[this.firstUnreadChannel]));

            if (firstUnreadElement.position().top + firstUnreadElement.height() < 0) {
                showTopUnread = true;
            }
        }

        if (this.lastUnreadChannel) {
            var lastUnreadElement = $(ReactDOM.findDOMNode(this.refs[this.lastUnreadChannel]));

            if (lastUnreadElement.position().top > container.height()) {
                showBottomUnread = true;
            }
        }

        this.setState({
            showTopUnread,
            showBottomUnread
        });
    }

    handleLeaveDirectChannel(channel) {
        if (!this.isLeaving.get(channel.id)) {
            this.isLeaving.set(channel.id, true);

            AsyncClient.savePreference(
                Constants.Preferences.CATEGORY_DIRECT_CHANNEL_SHOW,
                channel.teammate_id,
                'false',
                () => {
                    this.isLeaving.set(channel.id, false);
                },
                () => {
                    this.isLeaving.set(channel.id, false);
                }
            );

            this.setState(this.getStateFromStores());
        }

        if (channel.id === this.state.activeId) {
            browserHistory.push(TeamStore.getCurrentTeamUrl() + '/channels/town-square');
        }
    }

    sortChannelsByDisplayName(a, b) {
        return a.display_name.localeCompare(b.display_name);
    }

    sortMembersByUsername(a, b) {
        return a.username.localeCompare(b.username);
    }

    showMoreChannelsModal() {
        // manually show the modal because using data-toggle messes with keyboard focus when the modal is dismissed
        $('#more_channels').modal({'data-channeltype': 'O'}).modal('show');
    }

    showNewChannelModal(type) {
        this.setState({newChannelModalType: type});
    }
    hideNewChannelModal() {
        this.setState({newChannelModalType: ''});
    }

    showMoreDirectChannelsModal() {
        this.setState({showDirectChannelsModal: true});
    }
    hideMoreDirectChannelsModal() {
        this.setState({showDirectChannelsModal: false});
    }

    createTutorialTip() {
        const screens = [];

        screens.push(
            <div>
                <FormattedHTMLMessage
                    id='sidebar.tutorialScreen1'
                    defaultMessage='<h4>Channels</h4><p><strong>Channels</strong> organize conversations across different topics. They’re open to everyone on your team. To send private communications use <strong>Direct Messages</strong> for a single person or <strong>Private Groups</strong> for multiple people.</p>'
                />
            </div>
        );

        screens.push(
            <div>
                <FormattedHTMLMessage
                    id='sidebar.tutorialScreen2'
                    defaultMessage='<h4>"Town Square" and "Off-Topic" channels</h4>
                    <p>Here are two public channels to start:</p>
                    <p><strong>Town Square</strong> is a place for team-wide communication. Everyone in your team is a member of this channel.</p>
                    <p><strong>Off-Topic</strong> is a place for fun and humor outside of work-related channels. You and your team can decide what other channels to create.</p>'
                />
            </div>
        );

        screens.push(
            <div>
                <FormattedHTMLMessage
                    id='sidebar.tutorialScreen3'
                    defaultMessage='<h4>Creating and Joining Channels</h4>
                    <p>Click <strong>"More..."</strong> to create a new channel or join an existing one.</p>
                    <p>You can also create a new channel or private group by clicking the <strong>"+" symbol</strong> next to the channel or private group header.</p>'
                />
            </div>
        );

        return (
            <TutorialTip
                placement='right'
                screens={screens}
                overlayClass='tip-overlay--sidebar'
            />
        );
    }

    createMemberElement(member) {
        const userStatus = UserStore.getStatus(member.id);
        var status = null;
        var statusIcon = '';
        if (userStatus === 'online') {
            statusIcon = Constants.ONLINE_ICON_SVG;
        } else if (userStatus === 'away') {
            statusIcon = Constants.AWAY_ICON_SVG;
        } else {
            statusIcon = Constants.OFFLINE_ICON_SVG;
        }
        status = (
            <span
                className='status'
                dangerouslySetInnerHTML={{__html: statusIcon}}
            />
        );

        return (
            <li
                key={member.id}
            >
                <a className='sidebar-channel'>
                    {status}
                    {Utils.displayUsername(member.id)}
                </a>
            </li>
        );
    }

    createChannelElement(channel, index, arr, handleClose) {
        const members = this.state.members;
        const activeId = this.state.activeId;
        const channelMember = members[channel.id];
        const unreadCount = this.state.unreadCounts[channel.id] || {msgs: 0, mentions: 0};
        let msgCount;

        let linkClass = '';
        if (channel.id === activeId) {
            linkClass = 'active';
        }

        let rowClass = 'sidebar-channel';

        var unread = false;
        if (channelMember) {
            msgCount = unreadCount.msgs + unreadCount.mentions;
            unread = msgCount > 0 || channelMember.mention_count > 0;
        }

        if (unread) {
            rowClass += ' unread-title';

            if (channel.id !== activeId) {
                if (!this.firstUnreadChannel) {
                    this.firstUnreadChannel = channel.name;
                }
                this.lastUnreadChannel = channel.name;
            }
        }

        var badge = null;
        if (channelMember) {
            if (unreadCount.mentions) {
                badge = <span className='badge pull-right small'>{unreadCount.mentions}</span>;
                this.badgesActive = true;
            }
        } else if (this.state.loadingDMChannel === index && channel.type === 'D') {
            badge = (
                <img
                    className='channel-loading-gif pull-right'
                    src={loadingGif}
                />
            );
        }

        if (msgCount > 0) {
            rowClass += ' has-badge';
        }

        // set up status icon for direct message channels
        var status = null;
        if (channel.type === 'D') {
            var statusIcon = '';
            if (channel.status === 'online') {
                statusIcon = Constants.ONLINE_ICON_SVG;
            } else if (channel.status === 'away') {
                statusIcon = Constants.AWAY_ICON_SVG;
            } else {
                statusIcon = Constants.OFFLINE_ICON_SVG;
            }
            status = (
                <span
                    className='status'
                    dangerouslySetInnerHTML={{__html: statusIcon}}
                />
            );
        }

        var icon = null;
        if (channel.type === 'O') {
            icon = <div className='status'><i className='fa fa-globe'></i></div>;
        } else if (channel.type === 'P') {
            icon = <div className='status'><i className='fa fa-lock'></i></div>;
        }

        let closeButton = null;
        const removeTooltip = (
            <Tooltip id='remove-dm-tooltip'>
                <FormattedMessage
                    id='sidebar.removeList'
                    defaultMessage='Remove from list'
                />
            </Tooltip>
        );
        if (handleClose && !badge) {
            closeButton = (
                <OverlayTrigger
                    delayShow={1000}
                    placement='top'
                    overlay={removeTooltip}
                >
                    <span
                        onClick={() => handleClose(channel)}
                        className='btn-close'
                    >
                        {'×'}
                    </span>
                </OverlayTrigger>
            );

            rowClass += ' has-close';
        }

        let tutorialTip = null;
        if (this.state.showTutorialTip && channel.name === Constants.DEFAULT_CHANNEL) {
            tutorialTip = this.createTutorialTip();
        }

        let link = '';
        if (channel.fake) {
            link = TeamStore.getCurrentTeamUrl() + '/channels/' + channel.name + '?fakechannel=' + encodeURIComponent(JSON.stringify(channel));
        } else {
            link = TeamStore.getCurrentTeamUrl() + '/channels/' + channel.name;
        }

        return (
            <li
                key={channel.name}
                ref={channel.name}
                className={linkClass}
            >
                <Link
                    to={link}
                    className={rowClass}
                >
                    {icon}
                    {status}
                    {channel.display_name}
                    {badge}
                    {closeButton}
                </Link>
                {tutorialTip}
            </li>
        );
    }
    render() {
        // Check if we have all info needed to render
        if (this.state.currentTeam == null || this.state.currentUser == null) {
            return (<div/>);
        }

        this.badgesActive = false;

        // keep track of the first and last unread channels so we can use them to set the unread indicators
        this.firstUnreadChannel = null;
        this.lastUnreadChannel = null;

        // create elements for all 3 types of channels
        let publicChannelItems;
        let privateChannelItems;
        let directMessageItems;
        const isGuest = Utils.isGuest(this.state.currentUser.roles);

        if (isGuest) {
            const members = this.state.channelMembers.sort(this.sortMembersByUsername);
            directMessageItems = members.map(this.createMemberElement);
        } else {
            publicChannelItems = this.state.publicChannels.map(this.createChannelElement);

            privateChannelItems = this.state.privateChannels.map(this.createChannelElement);

            directMessageItems = this.state.directChannels.map((channel, index, arr) => {
                return this.createChannelElement(channel, index, arr, this.handleLeaveDirectChannel);
            });
        }

        // update the favicon to show if there are any notifications
        var link = document.createElement('link');
        link.type = 'image/x-icon';
        link.rel = 'shortcut icon';
        link.id = 'favicon';
        if (this.badgesActive) {
            link.href = redFavicon;
        } else {
            link.href = favicon;
        }
        var head = document.getElementsByTagName('head')[0];
        var oldLink = document.getElementById('favicon');
        if (oldLink) {
            head.removeChild(oldLink);
        }
        head.appendChild(link);

        var directMessageMore = null;
        if (this.state.hiddenDirectChannelCount > 0) {
            directMessageMore = (
                <li key='more'>
                    <a
                        href='#'
                        onClick={this.showMoreDirectChannelsModal}
                    >
                        <FormattedMessage
                            id='sidebar.more'
                            defaultMessage='More ({count})'
                            values={{
                                count: this.state.hiddenDirectChannelCount
                            }}
                        />
                    </a>
                </li>
            );
        }

        let showChannelModal = false;
        if (this.state.newChannelModalType !== '') {
            showChannelModal = true;
        }

        const createChannelTootlip = (
            <Tooltip id='new-channel-tooltip' >
                <FormattedMessage
                    id='sidebar.createChannel'
                    defaultMessage='Create new channel'
                />
            </Tooltip>
        );
        const createGroupTootlip = (
            <Tooltip id='new-group-tooltip'>
                <FormattedMessage
                    id='sidebar.createGroup'
                    defaultMessage='Create new group'
                />
            </Tooltip>
        );

        const above = (
            <FormattedMessage
                id='sidebar.unreadAbove'
                defaultMessage='Unread post(s) above'
            />
        );

        const below = (
            <FormattedMessage
                id='sidebar.unreadBelow'
                defaultMessage='Unread post(s) below'
            />
        );

        if (isGuest) {
            return (
                <div
                    className='sidebar--left'
                    id='sidebar-left'
                >
                    <SidebarHeader
                        teamDisplayName={this.state.currentTeam.display_name}
                        teamName={this.state.currentTeam.name}
                        teamType={this.state.currentTeam.type}
                        currentUser={this.state.currentUser}
                    />
                    <div
                        ref='container'
                        className='nav-pills__container'
                        onScroll={this.onScroll}
                    >
                        <ul className='nav nav-pills nav-stacked'>
                            <li>
                                <h4>
                                    <FormattedMessage
                                        id='sidebar.members'
                                        defaultMessage='Channel Members'
                                    />
                                </h4>
                            </li>
                            {directMessageItems}
                        </ul>
                    </div>
                </div>
            );
        }

        return (
            <div
                className='sidebar--left'
                id='sidebar-left'
                key='sidebar-left'
            >
                <NewChannelFlow
                    show={showChannelModal}
                    channelType={this.state.newChannelModalType}
                    onModalDismissed={this.hideNewChannelModal}
                />
                <MoreDirectChannels
                    show={this.state.showDirectChannelsModal}
                    onModalDismissed={this.hideMoreDirectChannelsModal}
                />

                <SidebarHeader
                    teamDisplayName={this.state.currentTeam.display_name}
                    teamName={this.state.currentTeam.name}
                    teamType={this.state.currentTeam.type}
                    currentUser={this.state.currentUser}
                />

                <UnreadChannelIndicator
                    show={this.state.showTopUnread}
                    extraClass='nav-pills__unread-indicator-top'
                    text={above}
                />
                <UnreadChannelIndicator
                    show={this.state.showBottomUnread}
                    extraClass='nav-pills__unread-indicator-bottom'
                    text={below}
                />

                <div
                    ref='container'
                    className='nav-pills__container'
                    onScroll={this.onScroll}
                >
                    <ul className='nav nav-pills nav-stacked'>
                        <li>
                            <h4>
                                <FormattedMessage
                                    id='sidebar.channels'
                                    defaultMessage='Channels'
                                />
                                <OverlayTrigger
                                    delayShow={500}
                                    placement='top'
                                    overlay={createChannelTootlip}
                                >
                                <a
                                    className='add-channel-btn'
                                    href='#'
                                    onClick={this.showNewChannelModal.bind(this, 'O')}
                                >
                                    {'+'}
                                </a>
                                </OverlayTrigger>
                            </h4>
                        </li>
                        {publicChannelItems}
                        <li>
                            <a
                                href='#'
                                className='nav-more'
                                onClick={this.showMoreChannelsModal}
                            >
                                <FormattedMessage
                                    id='sidebar.moreElips'
                                    defaultMessage='More...'
                                />
                            </a>
                        </li>
                    </ul>

                    <ul className='nav nav-pills nav-stacked'>
                        <li>
                            <h4>
                                <FormattedMessage
                                    id='sidebar.pg'
                                    defaultMessage='Private Groups'
                                />
                                <OverlayTrigger
                                    delayShow={500}
                                    placement='top'
                                    overlay={createGroupTootlip}
                                >
                                <a
                                    className='add-channel-btn'
                                    href='#'
                                    onClick={this.showNewChannelModal.bind(this, 'P')}
                                >
                                    {'+'}
                                </a>
                                </OverlayTrigger>
                            </h4>
                        </li>
                        {privateChannelItems}
                    </ul>
                    <ul className='nav nav-pills nav-stacked'>
                        <li>
                            <h4>
                                <FormattedMessage
                                    id='sidebar.direct'
                                    defaultMessage='Direct Messages'
                                />
                            </h4>
                        </li>
                        {directMessageItems}
                        {directMessageMore}
                    </ul>
                </div>
            </div>
        );
    }
}

Sidebar.defaultProps = {
};
Sidebar.propTypes = {
};
