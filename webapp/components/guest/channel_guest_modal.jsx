// Copyright (c) 2016 ZBox, Spa. All Rights Reserved.
// See License.txt for license information.

import $ from 'jquery';
import React from 'react';
import ReactDOM from 'react-dom';
import {Modal} from 'react-bootstrap';

import ChannelStore from 'stores/channel_store.jsx';
import * as Client from 'utils/client.jsx';

import {FormattedMessage, FormattedHTMLMessage} from 'react-intl';

export default class ChannelGuestModal extends React.Component {
    constructor() {
        super();

        this.handleInvite = this.handleInvite.bind(this);
        this.removeInvite = this.removeInvite.bind(this);
        this.onHide = this.onHide.bind(this);
        this.onListenerChange = this.onListenerChange.bind(this);
        this.copyLink = this.copyLink.bind(this);
        this.state = {
            inviteError: null,
            copiedLink: false
        };
    }
    componentDidMount() {
        ChannelStore.addGuestUrlListener(this.onListenerChange);
    }
    componentWillUnmount() {
        ChannelStore.removeGuestUrlListener(this.onListenerChange);
    }
    onListenerChange() {
        this.setState({
            inviteError: null,
            copiedLink: false,
            guestUrl: ChannelStore.getGuestUrl()
        });
    }
    onShow() {
        this.setState({
            inviteError: null,
            copiedLink: false,
            guestUrl: ChannelStore.getGuestUrl()
        });
    }
    onHide() {
        const state = this.state;
        state.copiedLink = false;
        this.setState(state);
        this.props.onHide();
    }
    componentDidUpdate(prevProps) {
        if (this.props.show && !prevProps.show) {
            this.onShow();
        }
    }
    copyLink(e) {
        e.preventDefault();
        const copyTextarea = $(ReactDOM.findDOMNode(this.refs.textarea));
        const state = this.state;
        copyTextarea.select();

        try {
            var successful = document.execCommand('copy');
            state.copiedLink = !!successful;
        } catch (err) {
            state.copiedLink = false;
        }

        this.setState(state);
    }
    handleInvite(e) {
        e.preventDefault();
        const obj = {
            channel_id: this.props.channel.id
        };

        Client.addGuestInvite(
            obj,
            (data) => {
                const url = global.window.location.origin + '/guest_signup/?id=' + data.invite_id;
                this.setState({
                    inviteError: null,
                    copiedLink: false,
                    guestUrl: url
                });
                ChannelStore.setGuestUrl(url);
            },
            (err) => {
                this.setState({
                    inviteError: err.message,
                    copiedLink: false,
                    guestUrl: null
                });
            }
        );
    }
    removeInvite(e) {
        e.preventDefault();
        if (this.state.guestUrl) {
            const inviteId = this.state.guestUrl.split('?id=')[1];
            Client.removeGuestInvite(inviteId,
                () => {
                    this.setState({
                        inviteError: null,
                        copiedLink: false,
                        guestUrl: null
                    });
                    ChannelStore.setGuestUrl(null);
                },
                (err) => {
                    const state = this.state;
                    state.inviteError = err.message;
                    state.copiedLink = false;
                    this.setState(state);
                }
            );
        }
    }
    render() {
        let inviteError = null;
        if (this.state.inviteError) {
            inviteError = (<div className='form-group has-error'><label className='control-label'>{this.state.inviteError}</label></div>);
        }

        let content;
        let copyLink = null;
        let copyLinkConfirm = null;
        let removeLink = null;
        let createLink = null;
        if (this.state.guestUrl) {
            content = (
                <div>
                    <p>
                        <FormattedHTMLMessage
                            id='channel_guest_modal.link_help'
                            defaultMessage='Copy the link and send it to someone that you would like to join the channel as a guest.<br/><br/>
                            <strong>Important: </strong>Anyone with this link would be able to create a <strong>guest</strong> account an join the conversation on this channel only.<br/><br/>
                            To revoke the open invitation and remove all the member guests click on <strong>"Uninvite & Revoke"</strong>'
                        />
                        <br/>
                        <br/>
                    </p>
                    <textarea
                        className='form-control no-resize min-height'
                        readOnly='true'
                        ref='textarea'
                        value={this.state.guestUrl}
                    />
                </div>
            );

            if (document.queryCommandSupported('copy')) {
                copyLink = (
                    <button
                        data-copy-btn='true'
                        type='button'
                        className='btn btn-primary pull-left'
                        onClick={this.copyLink}
                    >
                        <FormattedMessage
                            id='channel_guest_modal.copy'
                            defaultMessage='Copy Link'
                        />
                    </button>
                );
            }

            removeLink = (
                <button
                    type='button'
                    className='btn btn-danger pull-left'
                    onClick={this.removeInvite}
                >
                    <FormattedMessage
                        id='channel_guest_modal.remove'
                        defaultMessage='Uninvite & Revoke'
                    />
                </button>
            );

            if (this.state.copiedLink) {
                copyLinkConfirm = (
                    <p className='alert alert-success alert--confirm'>
                        <i className='fa fa-check'></i>
                        <FormattedMessage
                            id='channel_guest_modal.clipboard'
                            defaultMessage=' Link copied to clipboard.'
                        />
                    </p>
                );
            }
        } else {
            content = (
                <div>
                    <p>
                        <FormattedHTMLMessage
                            id='channel_guest_modal.add_help'
                            defaultMessage='Create a link to invite Guests to join the conversation on this channel without making them part of the Team.<br/><br/>
                            <strong>Important: </strong>When this link is enabled any person with it will be able to create a <strong>guest</strong> account and join this channel, leaving it open to the outside world.<br />
                            Share this link only with those you need and when no longer necesary revoke it<br/><br/>
                            To revoke the open invitation and remove all the member guests use the <strong>"Uninvite & Revoke"</strong> option'
                        />
                        <br/>
                        <br/>
                    </p>
                </div>
            );

            createLink = (
                <button
                    type='button'
                    className='btn btn-primary pull-left'
                    onClick={this.handleInvite}
                >
                    <FormattedMessage
                        id='channel_guest_modal.add'
                        defaultMessage='Enable Guests'
                    />
                </button>
            );
        }

        return (
            <Modal
                show={this.props.show}
                onHide={this.props.onHide}
            >
                <Modal.Header closeButton={true}>
                    <Modal.Title>
                        <FormattedMessage
                            id='channel_guest_modal.title'
                            defaultMessage='Guests Invite Link for '
                        />
                        <span className='name'>{this.props.channel.display_name}</span>
                    </Modal.Title>
                </Modal.Header>
                <Modal.Body
                    ref='modalBody'
                >
                    {inviteError}
                    {content}
                </Modal.Body>
                <Modal.Footer>
                    <button
                        type='button'
                        className='btn btn-default'
                        onClick={this.onHide}
                    >
                        <FormattedMessage
                            id='channel_guest_modal.close'
                            defaultMessage='Close'
                        />
                    </button>
                    {createLink}
                    {removeLink}
                    {copyLink}
                    {copyLinkConfirm}
                </Modal.Footer>
            </Modal>
        );
    }
}

ChannelGuestModal.propTypes = {
    show: React.PropTypes.bool.isRequired,
    onHide: React.PropTypes.func.isRequired,
    channel: React.PropTypes.object.isRequired
};
