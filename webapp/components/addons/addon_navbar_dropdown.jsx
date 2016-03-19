/**
 * Created by enahum on 2/8/16.
 */

import $ from 'jquery';
import React from 'react';
import ReactDOM from 'react-dom';
import * as Utils from 'utils/utils.jsx';
import * as Client from 'utils/client.jsx';
import TeamStore from 'stores/team_store.jsx';

import Constants from 'utils/constants.jsx';

import {FormattedMessage} from 'react-intl';

function getStateFromStores() {
    return {currentTeam: TeamStore.getCurrent()};
}

export default class AddonNavbarDropdown extends React.Component {
    constructor(props) {
        super(props);
        this.blockToggle = false;

        this.handleLogoutClick = this.handleLogoutClick.bind(this);

        this.state = getStateFromStores();
    }

    handleLogoutClick(e) {
        e.preventDefault();
        Client.logout();
    }

    componentDidMount() {
        $(ReactDOM.findDOMNode(this.refs.dropdown)).on('hide.bs.dropdown', () => {
            this.blockToggle = true;
            setTimeout(() => {
                this.blockToggle = false;
            }, 100);
        });
    }

    componentWillUnmount() {
        $(ReactDOM.findDOMNode(this.refs.dropdown)).off('hide.bs.dropdown');
    }

    render() {
        return (
            <ul className='nav navbar-nav navbar-right'>
                <li
                    ref='dropdown'
                    className='dropdown'
                >
                    <a
                        href='#'
                        className='dropdown-toggle'
                        data-toggle='dropdown'
                        role='button'
                        aria-expanded='false'
                    >
                        <span
                            className='dropdown__icon'
                            dangerouslySetInnerHTML={{__html: Constants.MENU_ICON}}
                        />
                    </a>
                    <ul
                        className='dropdown-menu'
                        role='menu'
                    >
                        <li>
                            <a
                                href={Utils.getWindowLocationOrigin() + '/' + this.state.currentTeam.name}
                            >
                                <FormattedMessage
                                    id='addon.nav.switch'
                                    defaultMessage='Switch to {display_name}'
                                    values={{
                                        display_name: this.state.currentTeam.display_name
                                    }}
                                />
                            </a>
                        </li>
                        <li>
                            <a
                                href='#'
                                onClick={this.handleLogoutClick}
                            >
                                <FormattedMessage
                                    id='addon.nav.logout'
                                    defaultMessage='Logout'
                                />
                            </a>
                        </li>
                        <li className='divider'></li>
                        <li>
                            <a
                                target='_blank'
                                href='/static/help/help.html'
                            >
                                <FormattedMessage
                                    id='addon.nav.help'
                                    defaultMessage='Help'
                                />
                            </a>
                        </li>
                        <li>
                            <a
                                target='_blank'
                                href='/static/help/report_problem.html'
                            >
                                <FormattedMessage
                                    id='addon.nav.report'
                                    defaultMessage='Report a Problem'
                                />
                            </a>
                        </li>
                    </ul>
                </li>
            </ul>
        );
    }
}