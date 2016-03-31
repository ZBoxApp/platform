// Copyright (c) 2015 ZBox, Spa. All Rights Reserved.
// See License.txt for license information.

import $ from 'jquery';
import React from 'react';
import ReactDOM from 'react-dom';
import * as Client from 'utils/client.jsx';
import * as AsyncClient from 'utils/async_client.jsx';

import {injectIntl, intlShape, defineMessages, FormattedMessage, FormattedHTMLMessage} from 'react-intl';

const holders = defineMessages({
    saving: {
        id: 'admin.video_call.saving',
        defaultMessage: 'Saving Config...'
    }
});

class VideoCallSettings extends React.Component {
    constructor(props) {
        super(props);

        this.handleChange = this.handleChange.bind(this);
        this.handleSubmit = this.handleSubmit.bind(this);

        this.state = {
            Enable: this.props.config.TwilioSettings.Enable,
            saveNeeded: false,
            serverError: null
        };
    }

    handleChange(action) {
        const s = {saveNeeded: true, serverError: this.state.serverError};

        if (action === 'EnableTrue') {
            s.Enable = true;
        }

        if (action === 'EnableFalse') {
            s.Enable = false;
        }

        this.setState(s);
    }

    handleSubmit(e) {
        e.preventDefault();
        $('#save-button').button('loading');

        const config = this.props.config;
        config.TwilioSettings.Enable = ReactDOM.findDOMNode(this.refs.Enable).checked;
        config.TwilioSettings.AccountSid = ReactDOM.findDOMNode(this.refs.AccountSid).value.trim();
        config.TwilioSettings.ApiKey = ReactDOM.findDOMNode(this.refs.ApiKey).value.trim();
        config.TwilioSettings.ApiSecret = ReactDOM.findDOMNode(this.refs.ApiSecret).value.trim();
        config.TwilioSettings.ConfigurationProfileSid = ReactDOM.findDOMNode(this.refs.ConfigurationProfileSid).value.trim();

        Client.saveConfig(
            config,
            () => {
                AsyncClient.getConfig();
                this.setState({
                    serverError: null,
                    saveNeeded: false
                });
                $('#save-button').button('reset');
            },
            (err) => {
                this.setState({
                    serverError: err.message,
                    saveNeeded: true
                });
                $('#save-button').button('reset');
            }
        );
    }

    render() {
        const {formatMessage} = this.props.intl;

        let serverError = '';
        if (this.state.serverError) {
            serverError = <div className='form-group has-error'><label className='control-label'>{this.state.serverError}</label></div>;
        }

        var saveClass = 'btn';
        if (this.state.saveNeeded) {
            saveClass = 'btn btn-primary';
        }

        return (
            <div className='wrapper--fixed'>
                <h3>
                    <FormattedMessage
                        id='admin.video_call.title'
                        defaultMessage='Video Call Settings'
                    />
                </h3>
                <form
                    className='form-horizontal'
                    role='form'
                >

                    <div className='form-group'>
                        <label
                            className='control-label col-sm-4'
                            htmlFor='Enable'
                        >
                            <FormattedMessage
                                id='admin.twilio.enableTitle'
                                defaultMessage='Enable Video Calls with Twilio: '
                            />
                        </label>
                        <div className='col-sm-8'>
                            <label className='radio-inline'>
                                <input
                                    type='radio'
                                    name='Enable'
                                    value='true'
                                    ref='Enable'
                                    defaultChecked={this.props.config.TwilioSettings.Enable}
                                    onChange={this.handleChange.bind(this, 'EnableTrue')}
                                />
                                <FormattedMessage
                                    id='admin.twilio.true'
                                    defaultMessage='true'
                                />
                            </label>
                            <label className='radio-inline'>
                                <input
                                    type='radio'
                                    name='Enable'
                                    value='false'
                                    defaultChecked={!this.props.config.TwilioSettings.Enable}
                                    onChange={this.handleChange.bind(this, 'EnableFalse')}
                                />
                                <FormattedMessage
                                    id='admin.twilio.false'
                                    defaultMessage='false'
                                />
                            </label>
                            <p className='help-text'>
                                <FormattedHTMLMessage
                                    id='admin.twilio.enableDescription'
                                    defaultMessage='When true, ZBox Now! allows making video calls to other teammates using Twilio Service.<br /><br />To configure, set the "Account Sid", "ApiKey", "ApiSecret" and "ConfigurationProfileSid" fields to complete the options bellow.<br />To get them Sign up to your <a href="https://www.twilio.com/user/account/video" target="_blank">Twilio account</a>.'
                                />
                            </p>
                        </div>
                    </div>

                    <div className='form-group'>
                        <label
                            className='control-label col-sm-4'
                            htmlFor='AccountSid'
                        >
                            <FormattedMessage
                                id='admin.twilio.accountSIdTitle'
                                defaultMessage='Account Sid:'
                            />
                        </label>
                        <div className='col-sm-8'>
                            <input
                                type='text'
                                className='form-control'
                                id='AccountSid'
                                ref='AccountSid'
                                placeholder=''
                                defaultValue={this.props.config.TwilioSettings.AccountSid}
                                onChange={this.handleChange}
                                disabled={!this.state.Enable}
                            />
                            <p className='help-text'>
                                <FormattedMessage
                                    id='admin.twilio.accountSidDescription'
                                    defaultMessage='Set the Account Sid provided by Twilio.'
                                />
                            </p>
                        </div>
                    </div>

                    <div className='form-group'>
                        <label
                            className='control-label col-sm-4'
                            htmlFor='ApiKey'
                        >
                            <FormattedMessage
                                id='admin.twilio.apiKeyTitle'
                                defaultMessage='Api Key:'
                            />
                        </label>
                        <div className='col-sm-8'>
                            <input
                                type='text'
                                className='form-control'
                                id='ApiKey'
                                ref='ApiKey'
                                placeholder=''
                                defaultValue={this.props.config.TwilioSettings.ApiKey}
                                onChange={this.handleChange}
                                disabled={!this.state.Enable}
                            />
                            <p className='help-text'>
                                <FormattedMessage
                                    id='admin.twilio.apiKeyDescription'
                                    defaultMessage='Set the Api Key provided by Twilio.'
                                />
                            </p>
                        </div>
                    </div>

                    <div className='form-group'>
                        <label
                            className='control-label col-sm-4'
                            htmlFor='ApiSecret'
                        >
                            <FormattedMessage
                                id='admin.twilio.apiSecretTitle'
                                defaultMessage='Api Secret:'
                            />
                        </label>
                        <div className='col-sm-8'>
                            <input
                                type='text'
                                className='form-control'
                                id='ApiSecret'
                                ref='ApiSecret'
                                placeholder=''
                                defaultValue={this.props.config.TwilioSettings.ApiSecret}
                                onChange={this.handleChange}
                                disabled={!this.state.Enable}
                            />
                            <p className='help-text'>
                                <FormattedMessage
                                    id='admin.twilio.apiSecretDescription'
                                    defaultMessage='Set the Api Secret provided by Twilio.'
                                />
                            </p>
                        </div>
                    </div>

                    <div className='form-group'>
                        <label
                            className='control-label col-sm-4'
                            htmlFor='ConfigurationProfileSid'
                        >
                            <FormattedMessage
                                id='admin.twilio.configurationProfileSidTitle'
                                defaultMessage='Configuration Profile Sid:'
                            />
                        </label>
                        <div className='col-sm-8'>
                            <input
                                type='text'
                                className='form-control'
                                id='ConfigurationProfileSid'
                                ref='ConfigurationProfileSid'
                                placeholder=''
                                defaultValue={this.props.config.TwilioSettings.ConfigurationProfileSid}
                                onChange={this.handleChange}
                                disabled={!this.state.Enable}
                            />
                            <p className='help-text'>
                                <FormattedHTMLMessage
                                    id='admin.twilio.configurationProfileSidDescription'
                                    defaultMessage='Enter your Configuration Profile Sid provided by Twilio.<br />Configuration Profiles store configurable parameters for Video applications.'
                                />
                            </p>
                        </div>
                    </div>

                    <div className='form-group'>
                        <div className='col-sm-12'>
                            {serverError}
                            <button
                                disabled={!this.state.saveNeeded}
                                type='submit'
                                className={saveClass}
                                onClick={this.handleSubmit}
                                id='save-button'
                                data-loading-text={'<span class=\'glyphicon glyphicon-refresh glyphicon-refresh-animate\'></span> ' + formatMessage(holders.saving)}
                            >
                                <FormattedMessage
                                    id='admin.video_call.save'
                                    defaultMessage='Save'
                                />
                            </button>
                        </div>
                    </div>

                </form>
            </div>
        );
    }
}

VideoCallSettings.propTypes = {
    intl: intlShape.isRequired,
    config: React.PropTypes.object
};

export default injectIntl(VideoCallSettings);