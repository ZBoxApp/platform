// Copyright (c) 2015 Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

import $ from 'jquery';
import 'bootstrap';
import NavbarSearchBox from './search_bar.jsx';
import MessageWrapper from './message_wrapper.jsx';
import PopoverListMembers from './popover_list_members.jsx';
import EditChannelHeaderModal from './edit_channel_header_modal.jsx';
import EditChannelPurposeModal from './edit_channel_purpose_modal.jsx';
import ChannelInfoModal from './channel_info_modal.jsx';
import ChannelInviteModal from './channel_invite_modal.jsx';
import ChannelMembersModal from './channel_members_modal.jsx';
import ChannelNotificationsModal from './channel_notifications_modal.jsx';
import ChannelGuestModal from './guest/channel_guest_modal.jsx';
import DeleteChannelModal from './delete_channel_modal.jsx';
import RenameChannelModal from './rename_channel_modal.jsx';
import ToggleModalButton from './toggle_modal_button.jsx';

import ChannelStore from 'stores/channel_store.jsx';
import UserStore from 'stores/user_store.jsx';
import SearchStore from 'stores/search_store.jsx';
import PreferenceStore from 'stores/preference_store.jsx';

import AppDispatcher from '../dispatcher/app_dispatcher.jsx';
import * as Utils from 'utils/utils.jsx';
import * as TextFormatting from 'utils/text_formatting.jsx';
import * as AsyncClient from 'utils/async_client.jsx';
import * as Client from 'utils/client.jsx';
import Constants from 'utils/constants.jsx';

import {FormattedMessage} from 'react-intl';
import {browserHistory} from 'react-router';

const ActionTypes = Constants.ActionTypes;

import {Tooltip, OverlayTrigger, Popover} from 'react-bootstrap';

import React from 'react';

export default class ChannelHeader extends React.Component {
    constructor(props) {
        super(props);

        this.onListenerChange = this.onListenerChange.bind(this);
        this.handleLeave = this.handleLeave.bind(this);
        this.searchMentions = this.searchMentions.bind(this);
        this.showRenameChannelModal = this.showRenameChannelModal.bind(this);
        this.hideRenameChannelModal = this.hideRenameChannelModal.bind(this);
        this.initialGuestUrl = this.initialGuestUrl.bind(this);
        this.makeCall = this.makeCall.bind(this);

        const state = this.getStateFromStores();
        state.showEditChannelPurposeModal = false;
        state.showMembersModal = false;
        state.showRenameChannelModal = false;
        this.state = state;
    }
    getStateFromStores() {
        const extraInfo = ChannelStore.getExtraInfo(this.props.channelId);
        const channel = ChannelStore.get(this.props.channelId);
        const currentUser = UserStore.getCurrentUser();
        let contact = null;

        if (currentUser) {
            if (channel.type === 'D' && extraInfo.members.length > 1) {
                if (extraInfo.members[0].id === currentUser.id) {
                    contact = extraInfo.members[1];
                } else {
                    contact = extraInfo.members[0];
                }
            }
        }

        return {
            channel,
            memberChannel: ChannelStore.getMember(this.props.channelId),
            guestUrl: ChannelStore.getGuestUrl(),
            users: extraInfo.members,
            userCount: extraInfo.member_count,
            searchVisible: SearchStore.getSearchResults() !== null,
            currentUser,
            contact
        };
    }
    validState() {
        if (!this.state.channel ||
            !this.state.memberChannel ||
            !this.state.users ||
            !this.state.userCount ||
            !this.state.currentUser) {
            return false;
        }
        return true;
    }
    componentDidMount() {
        this.initialGuestUrl();
        ChannelStore.addChangeListener(this.onListenerChange);
        ChannelStore.addExtraInfoChangeListener(this.onListenerChange);
        SearchStore.addSearchChangeListener(this.onListenerChange);
        PreferenceStore.addChangeListener(this.onListenerChange);
        UserStore.addChangeListener(this.onListenerChange);
        UserStore.addStatusesChangeListener(this.onListenerChange);
        $('.sidebar--left .dropdown-menu').perfectScrollbar();
    }
    componentWillUnmount() {
        ChannelStore.removeChangeListener(this.onListenerChange);
        ChannelStore.removeExtraInfoChangeListener(this.onListenerChange);
        SearchStore.removeSearchChangeListener(this.onListenerChange);
        PreferenceStore.removeChangeListener(this.onListenerChange);
        UserStore.removeChangeListener(this.onListenerChange);
        UserStore.removeStatusesChangeListener(this.onListenerChange);
    }
    initialGuestUrl() {
        Client.channelGuestInvite(ChannelStore.getCurrentId(),
            (data) => {
                let url = null;
                if (data) {
                    url = global.window.location.origin + '/guest_signup/?id=' + data.invite_id;
                }
                ChannelStore.setGuestUrl(url);
            }
        );
    }
    onListenerChange() {
        const newState = this.getStateFromStores();
        if (!Utils.areObjectsEqual(newState, this.state)) {
            this.setState(newState);
            this.initialGuestUrl();
        }
        $('.channel-header__info .description').popover({placement: 'bottom', trigger: 'hover', html: true, delay: {show: 500, hide: 500}});
    }
    handleLeave() {
        Client.leaveChannel(this.state.channel.id,
            () => {
                AppDispatcher.handleViewAction({
                    type: ActionTypes.LEAVE_CHANNEL,
                    id: this.state.channel.id
                });

                const townsquare = ChannelStore.getByName('town-square');
                browserHistory.push(Utils.getTeamURLNoOriginFromAddressBar() + '/channels/' + townsquare.name);
            },
            (err) => {
                AsyncClient.dispatchError(err, 'handleLeave');
            }
        );
    }
    searchMentions(e) {
        e.preventDefault();

        const user = this.state.currentUser;

        let terms = '';
        if (user.notify_props && user.notify_props.mention_keys) {
            const termKeys = UserStore.getMentionKeys(user.id);

            if (user.notify_props.all === 'true' && termKeys.indexOf('@all') !== -1) {
                termKeys.splice(termKeys.indexOf('@all'), 1);
            }

            if (user.notify_props.channel === 'true' && termKeys.indexOf('@channel') !== -1) {
                termKeys.splice(termKeys.indexOf('@channel'), 1);
            }
            terms = termKeys.join(' ');
        }

        AppDispatcher.handleServerAction({
            type: ActionTypes.RECEIVED_SEARCH_TERM,
            term: terms,
            do_search: true,
            is_mention_search: true
        });
    }
    showRenameChannelModal(e) {
        e.preventDefault();

        this.setState({
            showRenameChannelModal: true
        });
    }
    hideRenameChannelModal() {
        this.setState({
            showRenameChannelModal: false
        });
    }
    makeCall(data, isOnline) {
        if (isOnline) {
            GlobalActions.makeVideoCall(data.contact.id);
        }
    }
    render() {
        if (!this.validState()) {
            return null;
        }

        const channel = this.state.channel;
        const recentMentionsTooltip = (
            <Tooltip id='recentMentionsTooltip'>
                <FormattedMessage
                    id='channel_header.recentMentions'
                    defaultMessage='Recent Mentions'
                />
            </Tooltip>
        );
        const popoverContent = (
            <Popover
                id='hader-popover'
                bStyle='info'
                bSize='large'
                placement='bottom'
                className='description'
                onMouseOver={() => this.refs.headerOverlay.show()}
                onMouseOut={() => this.refs.headerOverlay.hide()}
            >
                <MessageWrapper
                    message={channel.header}
                />
            </Popover>
        );
        let channelTitle = channel.display_name;
        let makeCall;
        const roles = this.state.currentUser.roles;
        const isAdmin = Utils.isAdmin(roles) || Utils.isAdmin(roles);
        const isGuest = Utils.isGuest(roles) || Utils.isGuest(roles);
        const isDirect = (this.state.channel.type === 'D');

        if (isDirect) {
            const contact = this.state.contact;
            const userMedia = navigator.getUserMedia || navigator.webkitGetUserMedia || navigator.mozGetUserMedia;

            channelTitle = Utils.displayUsername(contact.id);

            if (global.window.mm_config.EnableTwilio === 'true' && userMedia) {
                const isOffline = UserStore.getStatus(contact.id) === 'offline';
                let circleClass = '';
                let offlineClass = 'on';
                if (isOffline) {
                    circleClass = 'offline';
                    offlineClass = 'off';
                }

                makeCall = (
                    <div className='video-call__header'>
                        <a
                            href='#'
                            onClick={() => this.makeCall(this.state, !isOffline)}
                            disabled={isOffline}
                        >
                            <svg
                                id='video-btn'
                                xmlns='http://www.w3.org/2000/svg'
                            >
                                <circle
                                    className={circleClass}
                                    cx='16'
                                    cy='16'
                                    r='18'
                                >
                                    <title>
                                        <FormattedMessage
                                            id='channel_header.make_video_call'
                                            defaultMessage='Make Video Call'
                                        />
                                    </title>
                                </circle>
                                <path
                                    className={offlineClass}
                                    transform='scale(0.4), translate(17,16)'
                                    d='M40 8H15.64l8 8H28v4.36l1.13 1.13L36 16v12.36l7.97 7.97L44 36V12c0-2.21-1.79-4-4-4zM4.55 2L2 4.55l4.01 4.01C4.81 9.24 4 10.52 4 12v24c0 2.21 1.79 4 4 4h29.45l4 4L44 41.46 4.55 2zM12 16h1.45L28 30.55V32H12V16z'
                                    fill='white'
                                />
                                <path
                                    className='off'
                                    transform='scale(0.4), translate(17,16)'
                                    d='M40 8H8c-2.21 0-4 1.79-4 4v24c0 2.21 1.79 4 4 4h32c2.21 0 4-1.79 4-4V12c0-2.21-1.79-4-4-4zm-4 24l-8-6.4V32H12V16h16v6.4l8-6.4v16z'
                                    fill='white'
                                />
                            </svg>
                        </a>
                    </div>
                );
            }
        }

        let channelTerm = (
            <FormattedMessage
                id='channel_header.channel'
                defaultMessage='Channel'
            />
        );
        if (channel.type === Constants.PRIVATE_CHANNEL) {
            channelTerm = (
                <FormattedMessage
                    id='channel_header.group'
                    defaultMessage='Group'
                />
            );
        }

        let popoverListMembers;
        if (!isDirect && !isGuest) {
            popoverListMembers = (
                <PopoverListMembers
                    channel={channel}
                    members={this.state.users}
                    memberCount={this.state.userCount}
                />
            );
        }

        const dropdownContents = [];
        if (isDirect) {
            dropdownContents.push(
                <li
                    key='edit_header_direct'
                    role='presentation'
                >
                    <ToggleModalButton
                        role='menuitem'
                        dialogType={EditChannelHeaderModal}
                        dialogProps={{channel}}
                    >
                        <FormattedMessage
                            id='channel_header.channelHeader'
                            defaultMessage='Set Channel Header...'
                        />
                    </ToggleModalButton>
                </li>
            );
        } else {
            dropdownContents.push(
                <li
                    key='view_info'
                    role='presentation'
                >
                    <ToggleModalButton
                        role='menuitem'
                        dialogType={ChannelInfoModal}
                        dialogProps={{channel}}
                    >
                        <FormattedMessage
                            id='channel_header.viewInfo'
                            defaultMessage='View Info'
                        />
                    </ToggleModalButton>
                </li>
            );

            if (!ChannelStore.isDefault(channel) && !isGuest) {
                dropdownContents.push(
                    <li
                        key='add_members'
                        role='presentation'
                    >
                        <ToggleModalButton
                            role='menuitem'
                            dialogType={ChannelInviteModal}
                            dialogProps={{channel, currentUser: this.state.currentUser}}
                        >
                            <FormattedMessage
                                id='chanel_header.addMembers'
                                defaultMessage='Add Members'
                            />
                        </ToggleModalButton>
                    </li>
                );

                if (isAdmin) {
                    dropdownContents.push(
                        <li
                            key='manage_members'
                            role='presentation'
                        >
                            <a
                                role='menuitem'
                                href='#'
                                onClick={() => this.setState({showMembersModal: true})}
                            >
                                <FormattedMessage
                                    id='channel_header.manageMembers'
                                    defaultMessage='Manage Members'
                                />
                            </a>
                        </li>
                    );
                }
            }

            if (!isGuest) {
                dropdownContents.push(
                    <li
                        key='set_channel_header'
                        role='presentation'
                    >
                        <ToggleModalButton
                            role='menuitem'
                            dialogType={EditChannelHeaderModal}
                            dialogProps={{channel}}
                        >
                            <FormattedMessage
                                id='channel_header.setHeader'
                                defaultMessage='Set {term} Header...'
                                values={{
                                    term: (channelTerm)
                                }}
                            />
                        </ToggleModalButton>
                    </li>
                );
                dropdownContents.push(
                    <li
                        key='set_channel_purpose'
                        role='presentation'
                    >
                        <a
                            role='menuitem'
                            href='#'
                            onClick={() => this.setState({showEditChannelPurposeModal: true})}
                        >
                            <FormattedMessage
                                id='channel_header.setPurpose'
                                defaultMessage='Set {term} Purpose...'
                                values={{
                                    term: (channelTerm)
                                }}
                            />
                        </a>
                    </li>
                );
            }

            dropdownContents.push(
                <li
                    key='notification_preferences'
                    role='presentation'
                >
                    <ToggleModalButton
                        role='menuitem'
                        dialogType={ChannelNotificationsModal}
                        dialogProps={{
                            channel,
                            channelMember: this.state.memberChannel,
                            currentUser: this.state.currentUser
                        }}
                    >
                        <FormattedMessage
                            id='channel_header.notificationPreferences'
                            defaultMessage='Notification Preferences'
                        />
                    </ToggleModalButton>
                </li>
            );

            if (!isDirect && isAdmin) {
                dropdownContents.push(
                    <li
                        key='invite_guests'
                        role='presentation'
                    >
                        <ToggleModalButton
                            role='menuitem'
                            dialogType={ChannelGuestModal}
                            dialogProps={{channel}}
                        >
                            <FormattedMessage
                                id='channel_header.guests'
                                defaultMessage='Invite Guests'
                            />
                        </ToggleModalButton>
                    </li>
                );
            }

            if (isAdmin) {
                dropdownContents.push(
                    <li
                        key='rename_channel'
                        role='presentation'
                    >
                        <a
                            role='menuitem'
                            href='#'
                            onClick={this.showRenameChannelModal}
                        >
                            <FormattedMessage
                                id='channel_header.rename'
                                defaultMessage='Rename {term}...'
                                values={{
                                    term: (channelTerm)
                                }}
                            />
                        </a>
                    </li>
                );

                if (!ChannelStore.isDefault(channel)) {
                    dropdownContents.push(
                        <li
                            key='delete_channel'
                            role='presentation'
                        >
                            <ToggleModalButton
                                role='menuitem'
                                dialogType={DeleteChannelModal}
                                dialogProps={{channel}}
                            >
                                <FormattedMessage
                                    id='channel_header.delete'
                                    defaultMessage='Delete {term}...'
                                    values={{
                                        term: (channelTerm)
                                    }}
                                />
                            </ToggleModalButton>
                        </li>
                    );
                }
            }

            if (!ChannelStore.isDefault(channel) && !isGuest) {
                dropdownContents.push(
                    <li
                        key='leave_channel'
                        role='presentation'
                    >
                        <a
                            role='menuitem'
                            href='#'
                            onClick={this.handleLeave}
                        >
                            <FormattedMessage
                                id='channel_header.leave'
                                defaultMessage='Leave {term}'
                                values={{
                                    term: (channelTerm)
                                }}
                            />
                        </a>
                    </li>
                );
            }
        }

        let navbarSearch;
        let mentionsSearch;
        if (!isGuest) {
            navbarSearch = (<th className='search-bar__container'><NavbarSearchBox/></th>);

            mentionsSearch = (
                <th>
                    <div className='dropdown channel-header__links'>
                        <OverlayTrigger
                            delayShow={Constants.OVERLAY_TIME_DELAY}
                            placement='bottom'
                            overlay={recentMentionsTooltip}
                        >
                            <a
                                href='#'
                                type='button'
                                onClick={this.searchMentions}
                            >
                                {'@'}
                            </a>
                        </OverlayTrigger>
                    </div>
                </th>
            );
        }

        return (
            <div
                id='channel-header'
                className='channel-header'
            >
                <table className='channel-header alt'>
                    <tbody>
                        <tr>
                            <th>
                                <div className='channel-header__info'>
                                    {makeCall}
                                    <div className='dropdown'>
                                        <a
                                            href='#'
                                            className='dropdown-toggle theme'
                                            type='button'
                                            id='channel_header_dropdown'
                                            data-toggle='dropdown'
                                            aria-expanded='true'
                                        >
                                            <strong className='heading'>{channelTitle} </strong>
                                            <span className='glyphicon glyphicon-chevron-down header-dropdown__icon'/>
                                        </a>
                                        <ul
                                            className='dropdown-menu'
                                            role='menu'
                                            aria-labelledby='channel_header_dropdown'
                                        >
                                            {dropdownContents}
                                        </ul>
                                    </div>
                                    <OverlayTrigger
                                        trigger={'click'}
                                        placement='bottom'
                                        overlay={popoverContent}
                                        ref='headerOverlay'
                                    >
                                    <div
                                        onClick={TextFormatting.handleClick}
                                        className='description'
                                        dangerouslySetInnerHTML={{__html: TextFormatting.formatText(channel.header, {singleline: true, mentionHighlight: false})}}
                                    />
                                    </OverlayTrigger>
                                </div>
                            </th>
                            <th>
                                {popoverListMembers}
                            </th>
                            {navbarSearch}
                            {mentionsSearch}
                        </tr>
                    </tbody>
                </table>
                <EditChannelPurposeModal
                    show={this.state.showEditChannelPurposeModal}
                    onModalDismissed={() => this.setState({showEditChannelPurposeModal: false})}
                    channel={channel}
                />
                <ChannelMembersModal
                    show={this.state.showMembersModal}
                    onModalDismissed={() => this.setState({showMembersModal: false})}
                    channel={channel}
                />
                <RenameChannelModal
                    show={this.state.showRenameChannelModal}
                    onHide={this.hideRenameChannelModal}
                    channel={channel}
                />
            </div>
        );
    }
}

ChannelHeader.propTypes = {
    channelId: React.PropTypes.string.isRequired
};
