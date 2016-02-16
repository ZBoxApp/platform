// Copyright (c) 2015 Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

import * as Utils from '../utils/utils.jsx';
import * as client from '../utils/client.jsx';
import UserStore from '../stores/user_store.jsx';
import BrowserStore from '../stores/browser_store.jsx';
import Constants from '../utils/constants.jsx';

import {intlShape, injectIntl, defineMessages, FormattedMessage, FormattedHTMLMessage} from 'mm-intl';

const holders = defineMessages({
    required: {
        id: 'guest_signup.required',
        defaultMessage: 'This field is required'
    },
    validEmail: {
        id: 'guest_signup.validEmail',
        defaultMessage: 'Please enter a valid email address'
    },
    reserved: {
        id: 'guest_signup.reserved',
        defaultMessage: 'This username is reserved, please choose a new one.'
    },
    usernameLength: {
        id: 'guest_signup.usernameLength',
        defaultMessage: 'Username must begin with a letter, and contain between {min} to {max} lowercase characters made up of numbers, letters, and the symbols \'.\', \'-\' and \'_\'.'
    },
    passwordLength: {
        id: 'guest_signup.passwordLength',
        defaultMessage: 'Please enter at least {min} characters'
    }
});

class GuestSignup extends React.Component {
    constructor(props) {
        super(props);

        this.handleSubmit = this.handleSubmit.bind(this);

        var initialState = BrowserStore.getGlobalItem(this.props.hash);

        if (!initialState) {
            initialState = {};
            initialState.wizard = 'welcome';
            initialState.user = {};
            initialState.user.team_id = this.props.teamId;
            initialState.user.email = this.props.email;
            initialState.original_email = this.props.email;
        }

        this.state = initialState;
    }
    handleSubmit(e) {
        e.preventDefault();

        const {formatMessage} = this.props.intl;
        const providedEmail = ReactDOM.findDOMNode(this.refs.email).value.trim();
        if (!providedEmail) {
            this.setState({nameError: '', emailError: formatMessage(holders.required), passwordError: ''});
            return;
        }

        if (!Utils.isEmail(providedEmail)) {
            this.setState({nameError: '', emailError: formatMessage(holders.validEmail), passwordError: ''});
            return;
        }

        const providedUsername = ReactDOM.findDOMNode(this.refs.name).value.trim().toLowerCase();
        if (!providedUsername) {
            this.setState({nameError: formatMessage(holders.required), emailError: '', passwordError: '', serverError: ''});
            return;
        }

        const usernameError = Utils.isValidUsername(providedUsername);
        if (usernameError === 'Cannot use a reserved word as a username.') {
            this.setState({nameError: formatMessage(holders.reserved), emailError: '', passwordError: '', serverError: ''});
            return;
        } else if (usernameError) {
            this.setState({
                nameError: formatMessage(holders.usernameLength, {min: Constants.MIN_USERNAME_LENGTH, max: Constants.MAX_USERNAME_LENGTH}),
                emailError: '',
                passwordError: '',
                serverError: ''
            });
            return;
        }

        const user = {
            team_id: this.props.teamId,
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

        const data = JSON.parse(this.props.data);

        client.createGuest(user, this.props.data,
            () => {
                client.loginGuest(this.props.teamName, user.email, data.invite_id,
                    () => {
                        UserStore.setLastEmail(user.email);
                        if (this.props.hash > 0) {
                            BrowserStore.setGlobalItem(this.props.hash, JSON.stringify({wizard: 'finished'}));
                        }
                        window.location.href = '/' + this.props.teamName + '/guest/' + data.channel_name;
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
        if (this.state.wizard === 'finished') {
            return (
                <div>
                    <FormattedMessage
                        id='guest_signup.expired'
                        defaultMessage="You've already completed the signup process for this invitation."
                    />
                </div>
            );
        }

        // set up error labels
        var emailError = null;
        var emailHelpText = (
            <span className='help-block'>
                <FormattedMessage
                    id='guest_signup.emailHelp'
                    defaultMessage='Valid email required for sign-up'
                />
            </span>
        );
        var emailDivStyle = 'form-group';
        if (this.state.emailError) {
            emailError = <label className='control-label'>{this.state.emailError}</label>;
            emailHelpText = '';
            emailDivStyle += ' has-error';
        }

        var nameError = null;
        var nameHelpText = (
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
        var nameDivStyle = 'form-group';
        if (this.state.nameError) {
            nameError = <label className='control-label'>{this.state.nameError}</label>;
            nameHelpText = '';
            nameDivStyle += ' has-error';
        }

        var serverError = null;
        if (this.state.serverError) {
            serverError = (
                <div className={'form-group has-error'}>
                    <label className='control-label'>{this.state.serverError}</label>
                </div>
            );
        }

        // set up the email entry and hide it if an email was provided
        var yourEmailIs = '';
        if (this.state.user.email) {
            yourEmailIs = (
                <FormattedHTMLMessage
                    id='guest_signup.emailIs'
                    defaultMessage="Your email address is <strong>{email}</strong>. You'll use this address to sign in to {siteName}."
                    values={{
                        email: this.state.user.email,
                        siteName: global.window.mm_config.SiteName
                    }}
                />
            );
        }

        var emailContainerStyle = 'margin--extra';
        if (this.state.original_email) {
            emailContainerStyle = 'hidden';
        }

        var email = (
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
                        defaultValue={this.state.user.email}
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

        var emailSignup;
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
                <form>
                    <img
                        className='signup-team-logo'
                        src='/static/images/logo.png'
                    />
                    <h5 className='margin--less'>
                        <FormattedMessage
                            id='guest_signup.welcome'
                            defaultMessage='Welcome to:'
                        />
                    </h5>
                    <h2 className='signup-team__name'>{this.props.teamDisplayName}</h2>
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
        );
    }
}

GuestSignup.defaultProps = {
    teamName: '',
    hash: '',
    teamId: '',
    email: '',
    data: null,
    teamDisplayName: ''
};
GuestSignup.propTypes = {
    intl: intlShape.isRequired,
    teamName: React.PropTypes.string,
    hash: React.PropTypes.string,
    teamId: React.PropTypes.string,
    email: React.PropTypes.string,
    data: React.PropTypes.string,
    teamDisplayName: React.PropTypes.string
};

export default injectIntl(GuestSignup);