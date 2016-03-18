// Copyright (c) 2015 Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

import TeamMembersModal from './team_members_modal.jsx';
import ToggleModalButton from './toggle_modal_button.jsx';
import UserSettingsModal from './user_settings/user_settings_modal.jsx';
import UserStore from 'stores/user_store.jsx';
import * as GlobalActions from 'action_creators/global_actions.jsx';
import * as Utils from 'utils/utils.jsx';

import {FormattedMessage} from 'react-intl';
import {Link} from 'react-router';

import React from 'react';

export default class SidebarRightMenu extends React.Component {
    constructor(props) {
        super(props);

        this.state = {
            showUserSettingsModal: false
        };
    }

    render() {
        var teamLink = '';
        var inviteLink = '';
        var teamSettingsLink = '';
        var manageLink = '';
        var consoleLink = '';
        var currentUser = UserStore.getCurrentUser();
        var isAdmin = false;
        var isSystemAdmin = false;
        var isGuest = false;

        if (currentUser != null) {
            isAdmin = Utils.isAdmin(currentUser.roles);
            isSystemAdmin = Utils.isSystemAdmin(currentUser.roles);
            isGuest = Utils.isGuest(currentUser.roles);

            if (!isGuest) {
                inviteLink = (
                    <li>
                        <a
                            href='#'
                            onClick={GlobalActions.showInviteMemberModal}
                        >
                            <i className='fa fa-user'></i>
                            <FormattedMessage
                                id='sidebar_right_menu.inviteNew'
                                defaultMessage='Invite New Member'
                            />
                        </a>
                    </li>
                );
            }

            if (this.props.teamType === 'O' && !isGuest) {
                teamLink = (
                    <li>
                        <a
                            href='#'
                            onClick={GlobalActions.showGetTeamInviteLinkModal}
                        >
                            <i className='glyphicon glyphicon-link'></i>
                            <FormattedMessage
                                id='sidebar_right_menu.teamLink'
                                defaultMessage='Get Team Invite Link'
                            />
                        </a>
                    </li>
                );
            }
        }

        if (isAdmin) {
            teamSettingsLink = (
                <li>
                    <a
                        href='#'
                        data-toggle='modal'
                        data-target='#team_settings'
                    >
                        <i className='fa fa-globe'></i>
                        <FormattedMessage
                            id='sidebar_right_menu.teamSettings'
                            defaultMessage='Team Settings'
                        />
                    </a>
                </li>
            );
            manageLink = (
                <li>
                    <ToggleModalButton dialogType={TeamMembersModal}>
                        <i className='fa fa-users'></i>
                        <FormattedMessage
                            id='sidebar_right_menu.manageMembers'
                            defaultMessage='Manage Members'
                        />
                    </ToggleModalButton>
                </li>
            );
        }

        if (isSystemAdmin && !Utils.isMobile()) {
            consoleLink = (
                <li>
                    <Link
                        to={'/admin_console'}
                    >
                        <i className='fa fa-wrench'></i>
                        <FormattedMessage
                            id='sidebar_right_menu.console'
                            defaultMessage='System Console'
                        />
                    </Link>
                </li>
            );
        }

        var siteName = '';
        if (global.window.mm_config.SiteName != null) {
            siteName = global.window.mm_config.SiteName;
        }
        var teamDisplayName = siteName;
        if (this.props.teamDisplayName) {
            teamDisplayName = this.props.teamDisplayName;
        }

        let helpLink = null;
        if (global.window.mm_config.HelpLink) {
            helpLink = (
                <li>
                    <Link
                        target='_blank'
                        to={global.window.mm_config.HelpLink}
                    >
                        <i className='fa fa-question'></i>
                        <FormattedMessage
                            id='sidebar_right_menu.help'
                            defaultMessage='Help'
                        />
                    </Link>
                </li>
            );
        }

        let reportLink = null;
        if (global.window.mm_config.ReportAProblemLink) {
            reportLink = (
                <li>
                    <Link
                        target='_blank'
                        to={global.window.mm_config.ReportAProblemLink}
                    >
                        <i className='fa fa-phone'></i>
                        <FormattedMessage
                            id='sidebar_right_menu.report'
                            defaultMessage='Report a Problem'
                        />
                    </Link>
                </li>
            );
        }
        return (
            <div
                className='sidebar--menu'
                id='sidebar-menu'
            >
                <div className='team__header theme'>
                    <Link
                        className='team__name'
                        to='/channels/town-square'
                    >
                        {teamDisplayName}
                    </Link>
                </div>

                <div className='nav-pills__container'>
                    <ul className='nav nav-pills nav-stacked'>
                        <li>
                            <a
                                href='#'
                                onClick={() => this.setState({showUserSettingsModal: true})}
                            >
                                <i className='fa fa-cog'></i>
                                <FormattedMessage
                                    id='sidebar_right_menu.accountSettings'
                                    defaultMessage='Account Settings'
                                />
                            </a>
                        </li>
                        {teamSettingsLink}
                        {inviteLink}
                        {teamLink}
                        {manageLink}
                        {consoleLink}
                        <li>
                            <Link to={Utils.getTeamURLFromAddressBar() + '/logout'}>
                                <i className='fa fa-sign-out'></i>
                                <FormattedMessage
                                    id='sidebar_right_menu.logout'
                                    defaultMessage='Logout'
                                />
                            </Link>
                        </li>
                        <li className='divider'></li>
                        {helpLink}
                        {reportLink}
                    </ul>
                </div>
                <UserSettingsModal
                    show={this.state.showUserSettingsModal}
                    onModalDismissed={() => this.setState({showUserSettingsModal: false})}
                />
            </div>
        );
    }
}

SidebarRightMenu.propTypes = {
    teamType: React.PropTypes.string,
    teamDisplayName: React.PropTypes.string
};
