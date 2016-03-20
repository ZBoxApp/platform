/**
 * Created by enahum on 2/8/16.
 */

import $ from 'jquery';
import React from 'react';
import AddonNavbarDropdown from './addon_navbar_dropdown.jsx';
import UserStore from 'stores/user_store.jsx';

import {FormattedMessage} from 'react-intl';

export default class AddonSidebarHeader extends React.Component {
    constructor(props) {
        super(props);

        this.toggleDropdown = this.toggleDropdown.bind(this);

        this.state = {};
    }

    toggleDropdown(e) {
        e.preventDefault();

        if (this.refs.dropdown.blockToggle) {
            this.refs.dropdown.blockToggle = false;
            return;
        }

        $('.team__header').find('.dropdown-toggle').dropdown('toggle');
    }

    render() {
        var me = UserStore.getCurrentUser();
        var profilePicture = null;

        if (!me) {
            return null;
        }

        if (me.last_picture_update) {
            profilePicture = (
                <img
                    className='user__picture'
                    src={'/api/v1/users/' + me.id + '/image?time=' + me.update_at}
                />
            );
        }

        return (
            <div className='team__header theme'>
                <a
                    href='#'
                    onClick={this.toggleDropdown}
                >
                    {profilePicture}
                    <div className='header__info'>
                        <div className='user__name'>{'@' + me.username}</div>
                        <div className='team__name'>
                            <FormattedMessage
                                id='addon.sidebarHeader.systemConsole'
                                defaultMessage='Team Addons'
                            />
                        </div>
                    </div>
                </a>
                <AddonNavbarDropdown ref='dropdown'/>
            </div>
        );
    }
}