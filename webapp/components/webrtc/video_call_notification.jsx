/**
 * Created by enahum on 3/25/16.
 */

import ChannelStore from 'stores/channel_store.jsx';
import UserStore from 'stores/user_store.jsx';
import VideoCallStore from 'stores/video_call_store.jsx';
import * as GlobalActions from 'action_creators/global_actions.jsx';
import * as Websockets from 'action_creators/websocket_actions.jsx';
import * as Utils from 'utils/utils.jsx';
import Constants from 'utils/constants.jsx';

import React from 'react';

import {FormattedMessage, FormattedHTMLMessage} from 'react-intl';

export default class VideoCallNotification extends React.Component {
    constructor() {
        super();

        this.mounted = false;

        this.closeBar = this.closeBar.bind(this);
        this.onIncomingCall = this.onIncomingCall.bind(this);
        this.onCancelCall = this.onCancelCall.bind(this);
        this.handleClose = this.handleClose.bind(this);
        this.handleAnswer = this.handleAnswer.bind(this);

        this.state = {
            channelId: ChannelStore.getCurrentId(),
            userCalling: null,
            callerId: null,
            notSupported: false
        };
    }

    componentDidMount() {
        VideoCallStore.addIncomingCallListener(this.onIncomingCall);
        VideoCallStore.addCancelCallListener(this.onCancelCall);
        this.mounted = true;
    }

    componentWillUnmount() {
        VideoCallStore.removeIncomingCallListener(this.onIncomingCall);
        VideoCallStore.removeCancelCallListener(this.onCancelCall);
        this.mounted = false;
    }

    closeBar() {
        this.setState({
            userCalling: null,
            callerId: null,
            notSupported: false
        });
    }

    onIncomingCall(incoming) {
        if (this.mounted) {
            const userId = incoming.from_id;
            const userMedia = navigator.getUserMedia || navigator.webkitGetUserMedia || navigator.mozGetUserMedia;

            if (userMedia) {
                // here we should check if the user is already in a call
                if (!this.state.callerId) {
                    this.setState({
                        userCalling: Utils.displayUsername(userId),
                        callerId: userId,
                        notSupported: false
                    });
                }
            } else {
                Websockets.sendMessage({
                    channel_id: this.state.channelId,
                    action: Constants.SocketEvents.VIDEO_CALL_NOT_SUPPORTED,
                    props: {
                        from_id: userId,
                        to_id: UserStore.getCurrentId()
                    }
                });

                this.setState({
                    userCalling: Utils.displayUsername(userId),
                    callerId: userId,
                    notSupported: true
                });
            }
        }
    }

    onCancelCall() {
        this.closeBar();
    }

    handleAnswer(e) {
        if (e) {
            e.preventDefault();
        }

        if (this.state.callerId) {
            GlobalActions.answerVideoCall(this.state.callerId);

            Websockets.sendMessage({
                channel_id: this.state.channelId,
                action: Constants.SocketEvents.VIDEO_CALL_ANSWER,
                props: {
                    from_id: this.state.callerId,
                    to_id: UserStore.getCurrentId()
                }
            });

            this.closeBar();
        }
    }

    handleClose(e) {
        if (e) {
            e.preventDefault();
        }

        if (this.state.callerId && !this.state.notSupported) {
            Websockets.sendMessage({
                channel_id: this.state.channelId,
                action: Constants.SocketEvents.VIDEO_CALL_REJECT,
                props: {
                    from_id: this.state.callerId,
                    to_id: UserStore.getCurrentId()
                }
            });
        }

        this.closeBar();
    }

    render() {
        let msg;
        if (this.state.callerId) {
            if (this.state.notSupported) {
                msg = (
                    <FormattedMessage
                        id='video_call_notification.not_supported'
                        defaultMessage='{username} is calling you, but your client does not support video calls'
                        values={{
                            username: this.state.userCalling
                        }}
                    />
                );
            } else {
                const answerLink = (
                    <a
                        href='#'
                        className='text-answer'
                        onClick={this.handleAnswer}
                    >
                        <FormattedHTMLMessage
                            id='video_call_notification.answer_call'
                            defaultMessage='<strong>Answer</strong>'
                        />
                    </a>
                );

                const rejectLink = (
                    <a
                        href='#'
                        className='text-reject'
                        onClick={this.handleClose}
                    >
                        <FormattedHTMLMessage
                            id='video_call_notification.reject_call'
                            defaultMessage='<strong>Reject</strong>'
                        />
                    </a>
                );

                msg = (
                    <FormattedMessage
                        id='video_call_notification.incoming_call'
                        defaultMessage='{username} is calling you. {answer} | {reject}'
                        values={{
                            username: this.state.userCalling,
                            answer: (answerLink),
                            reject: (rejectLink)
                        }}
                    />
                );
            }

            return (
                <div className='error-bar'>
                    <span>
                        {msg}
                    </span>
                    <a
                        href='#'
                        className='error-bar__close'
                        onClick={this.handleClose}
                    >
                        &times;
                    </a>
                </div>
            );
        }

        return <div/>;
    }
}

export default VideoCallNotification;
