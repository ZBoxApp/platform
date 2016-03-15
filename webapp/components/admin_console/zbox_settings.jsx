// Copyright (c) 2015 ZBox, Spa. All Rights Reserved.
// See License.txt for license information.

import $ from 'jquery';
import React from 'react';
import ReactDOM from 'react-dom';
import * as Client from '../../utils/client.jsx';
import * as AsyncClient from '../../utils/async_client.jsx';

import {intlShape, injectIntl, defineMessages, FormattedMessage} from 'react-intl';

const holders = defineMessages({
    clientIdExample: {
        id: 'admin.zbox.clientIdExample',
        defaultMessage: 'Ex "jcuS8PuvcpGhpgHhlcpT1Mx42pnqMxQY"'
    },
    clientSecretExample: {
        id: 'admin.zbox.clientSecretExample',
        defaultMessage: 'Ex "jcuS8PuvcpGhpgHhlcpT1Mx42pnqMxQY"'
    },
    authExample: {
        id: 'admin.zbox.authExample',
        defaultMessage: 'Ex ""'
    },
    tokenExample: {
        id: 'admin.zbox.tokenExample',
        defaultMessage: 'Ex ""'
    },
    userExample: {
        id: 'admin.zbox.userExample',
        defaultMessage: 'Ex ""'
    },
    loginExample: {
        id: 'admin.zbox.loginExample',
        defaultMessage: 'Ex ""'
    },
    saving: {
        id: 'admin.zbox.saving',
        defaultMessage: 'Saving Config...'
    }
});

class ZBoxSettings extends React.Component {
    constructor(props) {
        super(props);

        this.handleChange = this.handleChange.bind(this);
        this.handleSubmit = this.handleSubmit.bind(this);

        this.state = {
            Enable: this.props.config.ZBoxSettings.Enable,
            saveNeeded: false,
            serverError: null
        };
    }

    handleChange(action) {
        var s = {saveNeeded: true, serverError: this.state.serverError};

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

        var config = this.props.config;
        config.ZBoxSettings.Enable = ReactDOM.findDOMNode(this.refs.Enable).checked;
        config.ZBoxSettings.Secret = ReactDOM.findDOMNode(this.refs.Secret).value.trim();
        config.ZBoxSettings.Id = ReactDOM.findDOMNode(this.refs.Id).value.trim();
        config.ZBoxSettings.AuthEndpoint = ReactDOM.findDOMNode(this.refs.AuthEndpoint).value.trim();
        config.ZBoxSettings.TokenEndpoint = ReactDOM.findDOMNode(this.refs.TokenEndpoint).value.trim();
        config.ZBoxSettings.UserApiEndpoint = ReactDOM.findDOMNode(this.refs.UserApiEndpoint).value.trim();
        config.ZBoxSettings.LoginEndPoint = ReactDOM.findDOMNode(this.refs.LoginEndPoint).value.trim();

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
        var serverError = '';
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
                        id='admin.zbox.title'
                        defaultMessage='ZBox Settings'
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
                                id='admin.zbox.enableTitle'
                                defaultMessage='Enable Login With ZBox: '
                            />
                        </label>
                        <div className='col-sm-8'>
                            <label className='radio-inline'>
                                <input
                                    type='radio'
                                    name='Enable'
                                    value='true'
                                    ref='Enable'
                                    defaultChecked={this.props.config.ZBoxSettings.Enable}
                                    onChange={this.handleChange.bind(this, 'EnableTrue')}
                                />
                                <FormattedMessage
                                    id='admin.zbox.true'
                                    defaultMessage='true'
                                />
                            </label>
                            <label className='radio-inline'>
                                <input
                                    type='radio'
                                    name='Enable'
                                    value='false'
                                    defaultChecked={!this.props.config.ZBoxSettings.Enable}
                                    onChange={this.handleChange.bind(this, 'EnableFalse')}
                                />
                                <FormattedMessage
                                    id='admin.zbox.false'
                                    defaultMessage='false'
                                />
                            </label>
                            <p className='help-text'>
                                <FormattedMessage
                                    id='admin.zbox.enableDescription'
                                    defaultMessage='When true, ZBox Now allows account signup and login using ZBox OAuth. To configure, set the "Id", "Secret" and the Endpoints fields to complete the options below.'
                                />
                            </p>
                        </div>
                    </div>

                    <div className='form-group'>
                        <label
                            className='control-label col-sm-4'
                            htmlFor='Id'
                        >
                            <FormattedMessage
                                id='admin.zbox.clientIdTitle'
                                defaultMessage='Client Id:'
                            />
                        </label>
                        <div className='col-sm-8'>
                            <input
                                type='text'
                                className='form-control'
                                id='Id'
                                ref='Id'
                                placeholder={formatMessage(holders.clientIdExample)}
                                defaultValue={this.props.config.ZBoxSettings.Id}
                                onChange={this.handleChange}
                                disabled={!this.state.Enable}
                            />
                            <p className='help-text'>
                                <FormattedMessage
                                    id='admin.zbox.clientIdDescription'
                                    defaultMessage='Set the client id provided by ZBoxOAuth2 service.'
                                />
                            </p>
                        </div>
                    </div>

                    <div className='form-group'>
                        <label
                            className='control-label col-sm-4'
                            htmlFor='Secret'
                        >
                            <FormattedMessage
                                id='admin.zbox.clientSecretTitle'
                                defaultMessage='Client Secret:'
                            />
                        </label>
                        <div className='col-sm-8'>
                            <input
                                type='text'
                                className='form-control'
                                id='Secret'
                                ref='Secret'
                                placeholder={formatMessage(holders.clientSecretExample)}
                                defaultValue={this.props.config.ZBoxSettings.Secret}
                                onChange={this.handleChange}
                                disabled={!this.state.Enable}
                            />
                            <p className='help-text'>
                                <FormattedMessage
                                    id='admin.zbox.clientSecretDescription'
                                    defaultMessage='Set the client secret provided by ZBoxOAuth2 service.'
                                />
                            </p>
                        </div>
                    </div>

                    <div className='form-group'>
                        <label
                            className='control-label col-sm-4'
                            htmlFor='AuthEndpoint'
                        >
                            <FormattedMessage
                                id='admin.zbox.authTitle'
                                defaultMessage='Auth Endpoint:'
                            />
                        </label>
                        <div className='col-sm-8'>
                            <input
                                type='text'
                                className='form-control'
                                id='AuthEndpoint'
                                ref='AuthEndpoint'
                                placeholder={formatMessage(holders.authExample)}
                                defaultValue={this.props.config.ZBoxSettings.AuthEndpoint}
                                onChange={this.handleChange}
                                disabled={!this.state.Enable}
                            />
                            <p className='help-text'>
                                <FormattedMessage
                                    id='admin.zbox.authDescription'
                                    defaultMessage='Enter <your-zboxoauth2-url>/oauth/authorize (example http://localhost:3000/oauth/authorize).  Make sure you use HTTP or HTTPS in your URLs as appropriate.'
                                />
                            </p>
                        </div>
                    </div>

                    <div className='form-group'>
                        <label
                            className='control-label col-sm-4'
                            htmlFor='TokenEndpoint'
                        >
                            <FormattedMessage
                                id='admin.zbox.tokenTitle'
                                defaultMessage='Token Endpoint:'
                            />
                        </label>
                        <div className='col-sm-8'>
                            <input
                                type='text'
                                className='form-control'
                                id='TokenEndpoint'
                                ref='TokenEndpoint'
                                placeholder={formatMessage(holders.tokenExample)}
                                defaultValue={this.props.config.ZBoxSettings.TokenEndpoint}
                                onChange={this.handleChange}
                                disabled={!this.state.Enable}
                            />
                            <p className='help-text'>
                                <FormattedMessage
                                    id='admin.zbox.tokenDescription'
                                    defaultMessage='Enter <your-zboxoauth2-url>/oauth/token.   Make sure you use HTTP or HTTPS in your URLs as appropriate.'
                                />
                            </p>
                        </div>
                    </div>

                    <div className='form-group'>
                        <label
                            className='control-label col-sm-4'
                            htmlFor='UserApiEndpoint'
                        >
                            <FormattedMessage
                                id='admin.zbox.userTitle'
                                defaultMessage='User API Endpoint:'
                            />
                        </label>
                        <div className='col-sm-8'>
                            <input
                                type='text'
                                className='form-control'
                                id='UserApiEndpoint'
                                ref='UserApiEndpoint'
                                placeholder={formatMessage(holders.userExample)}
                                defaultValue={this.props.config.ZBoxSettings.UserApiEndpoint}
                                onChange={this.handleChange}
                                disabled={!this.state.Enable}
                            />
                            <p className='help-text'>
                                <FormattedMessage
                                    id='admin.zbox.userDescription'
                                    defaultMessage='Enter <your-zboxoauth2-url>/me.  Make sure you use HTTP or HTTPS in your URLs as appropriate.'
                                />
                            </p>
                        </div>
                    </div>

                    <div className='form-group'>
                        <label
                            className='control-label col-sm-4'
                            htmlFor='LoginEndPoint'
                        >
                            <FormattedMessage
                                id='admin.zbox.loginTitle'
                                defaultMessage='Login Endpoint:'
                            />
                        </label>
                        <div className='col-sm-8'>
                            <input
                                type='text'
                                className='form-control'
                                id='LoginEndPoint'
                                ref='LoginEndPoint'
                                placeholder={formatMessage(holders.loginExample)}
                                defaultValue={this.props.config.ZBoxSettings.LoginEndPoint}
                                onChange={this.handleChange}
                                disabled={!this.state.Enable}
                            />
                            <p className='help-text'>
                                <FormattedMessage
                                    id='admin.zbox.loginDescription'
                                    defaultMessage='Enter <your-zboxoauth2-url>/login.  Make sure you use HTTP or HTTPS in your URLs as appropriate.'
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
                                    id='admin.zbox.save'
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

ZBoxSettings.propTypes = {
    intl: intlShape.isRequired,
    config: React.PropTypes.object
};

export default injectIntl(ZBoxSettings);