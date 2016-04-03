// Copyright (c) 2015 ZBox, Spa. All Rights Reserved.
// See License.txt for license information.

import $ from 'jquery';
import React from 'react';
import ReactDOM from 'react-dom';

import * as Utils from '../../utils/utils.jsx';
import * as Client from '../../utils/client.jsx';
import UserStore from '../../stores/user_store.jsx';
import Constants from '../../utils/constants.jsx';
import LoadingScreen from '../../components/loading_screen.jsx';

import {FormattedMessage, FormattedHTMLMessage} from 'react-intl';
import {browserHistory} from 'react-router';

import logoImage from 'images/logo.png';

class GuestSignup extends React.Component {
    constructor(props) {
        super(props);

        this.handleSubmit = this.handleSubmit.bind(this);
        this.inviteInfoRecieved = this.inviteInfoRecieved.bind(this);
        this.inviteInfoError = this.inviteInfoError.bind(this);

        this.state = {
            email: '',
            channelId: '',
            channelName: '',
            inviteId: '',
            teamDisplayName: '',
            teamName: '',
            teamId: '',
            serverError: null
        };
    }
    componentWillMount() {
        const inviteId = this.props.location.query.id;
        Client.guestInviteInfo(this.inviteInfoRecieved, this.inviteInfoError, inviteId);
        document.title = Utils.localizeMessage('guest_signup.title', 'Signup as a Guest');
    }
    inviteInfoRecieved(data) {
        if (!data) {
            return;
        }

        if (data.redirect_uri) {
            browserHistory.push(data.redirect_uri);
        } else {
            this.setState({
                inviteId: data.invite_id,
                channelId: data.channel_id,
                channelName: data.channel_name,
                teamDisplayName: data.team_display_name,
                teamName: data.team_name,
                teamId: data.team_id
            });
        }
    }
    inviteInfoError(err) {
        const invalidParams = {
            title: Utils.localizeMessage('error.guest_invite.title', 'Invitation not found'),
            message: err.message,
            link: '/',
            linkmessage: Utils.localizeMessage('error.not_found.link_message', 'Back to ZBox Now!')
        };

        browserHistory.push('/error?' + $.param(invalidParams));
    }
    handleSubmit(e) {
        e.preventDefault();

        const providedEmail = ReactDOM.findDOMNode(this.refs.email).value.trim();
        if (!providedEmail) {
            this.setState({
                nameError: '',
                emailError: (
                    <FormattedMessage
                        id='guest_signup.required'
                        defaultMessage='This field is required'
                    />),
                passwordError: ''
            });
            return;
        }

        if (!Utils.isEmail(providedEmail)) {
            this.setState({
                nameError: '',
                emailError: (
                    <FormattedMessage
                        id='guest_signup.validEmail'
                        defaultMessage='Please enter a valid email address'
                    />),
                passwordError: ''
            });
            return;
        }

        const providedUsername = ReactDOM.findDOMNode(this.refs.name).value.trim().toLowerCase();
        if (!providedUsername) {
            this.setState({
                nameError: (
                    <FormattedMessage
                        id='guest_signup.required'
                        defaultMessage='This field is required'
                    />),
                emailError: '',
                passwordError: '',
                serverError: ''
            });
            return;
        }

        const usernameError = Utils.isValidUsername(providedUsername);
        if (usernameError === 'Cannot use a reserved word as a username.') {
            this.setState({
                nameError: (
                    <FormattedMessage
                        id='guest_signup.reserved'
                        defaultMessage='This username is reserved, please choose a new one.'
                    />
                ),
                emailError: '',
                passwordError: '',
                serverError: ''
            });
            return;
        } else if (usernameError) {
            this.setState({
                nameError: (
                    <FormattedMessage
                        id='guest_signup.usernameLength'
                        defaultMessage="Username must begin with a letter, and contain between {min} to {max} lowercase characters made up of numbers, letters, and the symbols '.', '-' and '_'."
                        values={{
                            min: Constants.MIN_USERNAME_LENGTH,
                            max: Constants.MAX_USERNAME_LENGTH
                        }}
                    />
                ),
                emailError: '',
                passwordError: '',
                serverError: ''
            });
            return;
        }

        const user = {
            team_id: this.state.teamId,
            email: providedEmail,
            username: providedUsername,
            allow_marketing: true
        };

        this.setState({
            user,
            nameError: '',
            emailError: '',
            serverError: ''
        });

        const state = this.state;
        const data = {
            invite_id: state.inviteId,
            team_id: state.teamId,
            team_name: state.teamName,
            channel_id: state.channelId,
            channel_name: state.channelName
        };

        Client.createGuest(user, JSON.stringify(data),
            () => {
                Client.loginGuest(data.team_name, user.email, data.invite_id,
                    () => {
                        UserStore.setLastEmail(user.email);
                        browserHistory.push('/' + data.team_name + '/guest/' + data.channel_name);
                    },
                    (err) => {
                        this.setState({serverError: err.message});
                    }
                );
            },
            (err) => {
                this.setState({serverError: err.message});
            }
        );
    }
    render() {
        if (this.state.teamId === '' && !this.state.inviteError) {
            return (<LoadingScreen/>);
        }

        // set up error labels
        let emailError = null;
        let emailHelpText = (
            <span className='help-block'>
                <FormattedMessage
                    id='guest_signup.emailHelp'
                    defaultMessage='Valid email required for sign-up'
                />
            </span>
        );
        let emailDivStyle = 'form-group';
        if (this.state.emailError) {
            emailError = <label className='control-label'>{this.state.emailError}</label>;
            emailHelpText = '';
            emailDivStyle += ' has-error';
        }

        let nameError = null;
        let nameHelpText = (
            <span className='help-block'>
                <FormattedMessage
                    id='guest_signup.userHelp'
                    defaultMessage="Username must begin with a letter, and contain between {min} to {max} lowercase characters made up of numbers, letters, and the symbols '.', '-' and '_'"
                    values={{
                        min: Constants.MIN_USERNAME_LENGTH,
                        max: Constants.MAX_USERNAME_LENGTH
                    }}
                />
            </span>
        );
        let nameDivStyle = 'form-group';
        if (this.state.nameError) {
            nameError = <label className='control-label'>{this.state.nameError}</label>;
            nameHelpText = '';
            nameDivStyle += ' has-error';
        }

        let serverError = null;
        if (this.state.serverError) {
            serverError = (
                <div className={'form-group has-error'}>
                    <label className='control-label'>{this.state.serverError}</label>
                </div>
            );
        }

        // set up the email entry and hide it if an email was provided
        let yourEmailIs = '';
        if (this.state.email) {
            yourEmailIs = (
                <FormattedHTMLMessage
                    id='guest_signup.emailIs'
                    defaultMessage="Your email address is <strong>{email}</strong>. You'll use this address to sign in to {siteName}."
                    values={{
                        email: this.state.email,
                        siteName: global.window.mm_config.SiteName
                    }}
                />
            );
        }

        let emailContainerStyle = 'margin--extra';
        if (this.state.email) {
            emailContainerStyle = 'hidden';
        }

        const email = (
            <div className={emailContainerStyle}>
                <h5><strong>
                    <FormattedMessage
                        id='guest_signup.whatis'
                        defaultMessage="What's your email address?"
                    />
                </strong></h5>
                <div className={emailDivStyle}>
                    <input
                        type='email'
                        ref='email'
                        className='form-control'
                        defaultValue={this.state.email}
                        placeholder=''
                        maxLength='128'
                        autoFocus={true}
                        spellCheck='false'
                    />
                    {emailError}
                    {emailHelpText}
                </div>
            </div>
        );

        let emailSignup;
        if (global.window.mm_config.EnableSignUpWithEmail === 'true') {
            emailSignup = (
                <div>
                    <div className='inner__content'>
                        {email}
                        {yourEmailIs}
                        <div className='margin--extra'>
                            <h5><strong>
                                <FormattedMessage
                                    id='guest_signup.chooseUser'
                                    defaultMessage='Choose your username'
                                />
                            </strong></h5>
                            <div className={nameDivStyle}>
                                <input
                                    type='text'
                                    ref='name'
                                    className='form-control'
                                    placeholder=''
                                    maxLength={Constants.MAX_USERNAME_LENGTH}
                                    spellCheck='false'
                                />
                                {nameError}
                                {nameHelpText}
                            </div>
                        </div>
                    </div>
                    <p className='margin--extra'>
                        <button
                            type='submit'
                            onClick={this.handleSubmit}
                            className='btn-primary btn'
                        >
                            <FormattedMessage
                                id='guest_signup.create'
                                defaultMessage='Create Guest Account'
                            />
                        </button>
                    </p>
                </div>
            );
        }

        if (!emailSignup) {
            emailSignup = (
                <div>
                    <FormattedMessage
                        id='guest_signup.none'
                        defaultMessage='No user creation method has been enabled.  Please contact an administrator for access.'
                    />
                </div>
            );
        }

        return (
        <div>
            <div className='signup-header'>
                <a href='/'>
                    <span className='fa fa-chevron-left'/>
                    <FormattedMessage id='web.header.back'/>
                </a>
            </div>
            <div className='col-sm-12'>
                <div className='signup-team__container padding--less'>
                    <form>
                        <img
                            className='signup-team-logo'
                            src={logoImage}
                        />
                        <h5 className='margin--less'>
                            <FormattedMessage
                                id='signup_user_completed.welcome'
                                defaultMessage='Welcome to:'
                            />
                        </h5>
                        <h2 className='signup-team__name'>{this.state.teamDisplayName}</h2>
                        <h2 className='signup-team__subdomain'>
                            <FormattedMessage
                                id='guest_signup.onSite'
                                defaultMessage='on {siteName}'
                                values={{
                                    siteName: global.window.mm_config.SiteName
                                }}
                            />
                        </h2>
                        <h4 className='color--light'>
                            <FormattedMessage
                                id='guest_signup.lets'
                                defaultMessage="Let's create your account"
                            />
                        </h4>
                        {emailSignup}
                        {serverError}
                    </form>
                </div>
            </div>
        </div>
        );
    }
}

GuestSignup.defaultProps = {
};
GuestSignup.propTypes = {
    location: React.PropTypes.object,
    history: React.PropTypes.object
};

export default GuestSignup;