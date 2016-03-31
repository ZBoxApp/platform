/**
 * Created by enahum on 3/24/16.
 */

import $ from 'jquery';
import UserStore from 'stores/user_store.jsx';
import ChannelStore from 'stores/channel_store.jsx';
import VideoCallStore from 'stores/video_call_store.jsx';

import * as Client from 'utils/client.jsx';
import * as Utils from 'utils/utils.jsx';
import * as Websockets from 'action_creators/websocket_actions.jsx';

import SearchBox from '../search_bar.jsx';
import VideoCallHeader from './video_call_header.jsx';
import Connecting from './connecting_screen.jsx';

import AppDispatcher from 'dispatcher/app_dispatcher.jsx';
import Constants from 'utils/constants.jsx';

import {AccessManager} from 'twilio-common';
import Conversations from 'twilio-conversations';

import React from 'react';
import {injectIntl, intlShape, defineMessages, FormattedMessage} from 'react-intl';

const ActionTypes = Constants.ActionTypes;

const holders = defineMessages({
    mediaError: {
        id: 'video_call.mediaError',
        defaultMessage: 'Unable to access Camera and Microphone'
    },
    rejected: {
        id: 'video_call.rejected',
        defaultMessage: 'Your call has been rejected by '
    },
    offline: {
        id: 'video_call.offline',
        defaultMessage: ' is offline'
    },
    inProgress: {
        id: 'video_call.inProgress',
        defaultMessage: 'You have a video call in progress. Hangup before closing this window.'
    },
    unable: {
        id: 'video_call.unable',
        defaultMessage: 'Unable to create conversation'
    },
    tokenFailed: {
        id: 'video_call.tokenFailed',
        defaultMessage: 'We encountered an error while getting the video call token'
    },
    notSupported: {
        id: 'video_call.notSupported',
        defaultMessage: "{username}'s client does not support video calls"
    },
    cancelled: {
        id: 'video_call.cancelled',
        defaultMessage: ' cancelled the call'
    },
    failed: {
        id: 'video_call.failed',
        defaultMessage: 'There was a problem creating the video call'
    }
});

class VideoCall extends React.Component {
    constructor(props) {
        super(props);

        this.mounted = false;
        this.localMedia = null;
        this.accessManager = null;
        this.conversationsClient = null;
        this.conversation = null;
        this.outgoingInvitation = null;

        this.handleResize = this.handleResize.bind(this);
        this.handleClose = this.handleClose.bind(this);
        this.handlePlaceCall = this.handlePlaceCall.bind(this);
        this.handleCancelCall = this.handleCancelCall.bind(this);
        this.handleFailedCall = this.handleFailedCall.bind(this);

        this.onStatusChange = this.onStatusChange.bind(this);
        this.onCallRejected = this.onCallRejected.bind(this);
        this.onConnectCall = this.onConnectCall.bind(this);
        this.onCancelCall = this.onCancelCall.bind(this);
        this.onToggleAudio = this.onToggleAudio.bind(this);
        this.onToggleVideo = this.onToggleVideo.bind(this);
        this.onNotSupported = this.onNotSupported.bind(this);
        this.onFailed = this.onFailed.bind(this);

        this.previewVideo = this.previewVideo.bind(this);
        this.connected = this.connected.bind(this);
        this.conversationStarted = this.conversationStarted.bind(this);
        this.localMediaToMain = this.localMediaToMain.bind(this);
        this.mainMediaToLocal = this.mainMediaToLocal.bind(this);
        this.unregister = this.unregister.bind(this);
        this.close = this.close.bind(this);

        this.state = {
            windowWidth: Utils.windowWidth(),
            windowHeight: Utils.windowHeight(),
            channelId: ChannelStore.getCurrentId(),
            currentUser: UserStore.getCurrentUser(),
            localMediaLoaded: false,
            isPaused: false,
            isMuted: false,
            isCalling: false,
            callInProgress: false,
            error: null
        };
    }

    componentDidMount() {
        this.resize();
        window.addEventListener('resize', this.handleResize);

        VideoCallStore.addRejectedCallListener(this.onCallRejected);
        VideoCallStore.addCancelCallListener(this.onCancelCall);
        VideoCallStore.addConnectCallListener(this.onConnectCall);
        VideoCallStore.addNotSupportedCallListener(this.onNotSupported);
        VideoCallStore.addFailedCallListener(this.onFailed);

        UserStore.addStatusesChangeListener(this.onStatusChange);

        this.mounted = true;
        if (!Utils.isMobile()) {
            $('.sidebar--right .post-right__scroll').perfectScrollbar();
        }

        this.previewVideo(false);
    }

    componentWillUnmount() {
        window.removeEventListener('resize', this.handleResize);

        VideoCallStore.removeRejectedCallListener(this.onCallRejected);
        VideoCallStore.removeCancelCallListener(this.onCancelCall);
        VideoCallStore.removeConnectCallListener(this.onConnectCall);
        VideoCallStore.removeNotSupportedCallListener(this.onNotSupported);
        VideoCallStore.removeFailedCallListener(this.onFailed);

        UserStore.removeStatusesChangeListener(this.onStatusChange);

        this.mounted = false;
    }

    handleResize() {
        this.setState({
            windowWidth: Utils.windowWidth(),
            windowHeight: Utils.windowHeight()
        });
    }

    previewVideo(bypassCaller) {
        if ((this.props.isCaller || bypassCaller) && this.mounted) {
            if (this.localMedia) {
                this.setState({
                    localMediaLoaded: true
                });
            } else {
                this.localMedia = new Conversations.LocalMedia();
                Conversations.getUserMedia().then(
                    (mediaStream) => {
                        this.localMedia.addStream(mediaStream);
                        this.localMedia.attach('#main-video');
                        this.setState({
                            localMediaLoaded: true
                        });
                    },
                    () => {
                        this.setState({
                            error: Utils.localizeMessage(holders.mediaError.id, holders.mediaError.defaultMessage)
                        });
                    });
            }
        }
    }

    handlePlaceCall() {
        if (UserStore.getStatus(this.props.userId) !== 'offline') {
            this.setState({
                isCalling: true,
                callInProgress: false,
                error: null
            });

            Websockets.sendMessage({
                channel_id: this.state.channelId,
                action: Constants.SocketEvents.START_VIDEO_CALL,
                props: {
                    from_id: UserStore.getCurrentId(),
                    to_id: this.props.userId
                }
            });
        }
    }

    handleCancelCall() {
        Websockets.sendMessage({
            channel_id: this.state.channelId,
            action: Constants.SocketEvents.CANCEL_VIDEO_CALL,
            props: {
                from_id: UserStore.getCurrentId(),
                to_id: this.props.userId
            }
        });

        this.unregister();
    }

    handleFailedCall() {
        Websockets.sendMessage({
            channel_id: this.state.channelId,
            action: Constants.SocketEvents.VIDEO_CALL_FAILED,
            props: {
                from_id: UserStore.getCurrentId(),
                to_id: this.props.userId
            }
        });

        this.unregister();
    }

    onCancelCall() {
        const isAnswering = this.state.isAnswering;
        this.unregister();
        this.previewVideo(true);

        if (this.mounted && isAnswering) {
            this.setState({
                error: `${Utils.displayUsername(this.props.userId)} ${Utils.localizeMessage(holders.cancelled.id, holders.cancelled.defaultMessage)}`
            });
        }
    }

    handleClose() {
        if (this.conversation) {
            this.setState({
                error: Utils.localizeMessage(holders.inProgress.id, holders.inProgress.defaultMessage)
            });
        } else if (this.state.isCalling) {
            this.handleCancelCall();
            this.close();
        } else {
            this.unregister();
            this.close();
        }
    }

    onCallRejected() {
        let error = null;

        if (this.state.isCalling) {
            error = `${Utils.localizeMessage(holders.rejected.id, holders.rejected.defaultMessage)} ${Utils.displayUsername(this.props.userId)}`;
        }

        this.setState({
            isCalling: false,
            callInProgress: false,
            error
        });
    }

    onStatusChange() {
        const status = UserStore.getStatus(this.props.userId);

        if (status === 'offline') {
            const error = `${Utils.displayUsername(this.props.userId)} ${Utils.localizeMessage(holders.offline.id, holders.offline.defaultMessage)}`;

            if (this.state.isCalling) {
                this.setState({
                    isCalling: false,
                    callInProgress: false,
                    error
                });
            } else {
                this.setState({
                    error
                });
            }
        } else if (status !== 'offline' && this.state.error) {
            this.setState({
                error: null
            });
        }
    }

    onConnectCall() {
        const self = this;
        Client.getTwilioToken(
            (data) => {
                self.accessManager = new AccessManager(data.token);
                self.conversationsClient = new Conversations.Client(self.accessManager);

                self.conversationsClient.listen().then(
                    self.connected,
                    (error) => {
                        console.error(Utils.localizeMessage(holders.unable.id, holders.unable.defaultMessage), error); //eslint-disable-line no-console

                        self.handleFailedCall();
                    }
                );
            },
            (error) => {
                console.error(Utils.localizeMessage(holders.tokenFailed.id, holders.tokenFailed.defaultMessage), error); //eslint-disable-line no-console

                self.handleFailedCall();
            }
        );
    }

    onToggleAudio() {
        this.conversation.localMedia.mute(!this.state.isMuted);
        this.setState({
            isMuted: !this.state.isMuted
        });
    }

    onToggleVideo() {
        this.conversation.localMedia.pause(!this.state.isPaused);
        this.setState({
            isPaused: !this.state.isPaused
        });
    }

    onNotSupported() {
        if (this.mounted) {
            this.setState({
                error: this.props.intl.formatMessage(holders.notSupported, {username: Utils.displayUsername(this.props.userId)}),
                callInProgress: false,
                isCalling: false,
                isAnswering: false
            });

            this.conversationsClient = null;
            this.conversation = null;
            this.accessManager = null;
        }
    }

    onFailed() {
        this.setState({
            isCalling: false,
            callInProgress: false,
            error: Utils.localizeMessage(holders.failed.id, holders.failed.defaultMessage)
        });

        this.previewVideo(true);
    }

    unregister() {
        if (this.conversation) {
            this.conversation.disconnect();
        } else if (this.outgoingInvitation) {
            this.outgoingInvitation.cancel();
        }

        if (this.conversationsClient) {
            this.conversationsClient.unlisten();
        }

        this.conversationsClient = null;
        this.outgoingInvitation = null;
        this.conversation = null;
        this.accessManager = null;

        if (this.mounted) {
            this.setState({
                error: null,
                callInProgress: false,
                isCalling: false,
                isAnswering: false
            });
        }
    }

    close() {
        if (this.localMedia) {
            this.localMedia.stop();
            this.localMedia = null;
        }

        AppDispatcher.handleServerAction({
            type: ActionTypes.RECEIVED_SEARCH,
            results: null
        });

        AppDispatcher.handleServerAction({
            type: ActionTypes.RECEIVED_POST_SELECTED,
            postId: null
        });
    }

    connected() {
        if (this.props.isCaller || this.state.isCalling) {
            const options = {};

            if (this.localMedia) {
                options.localMedia = this.localMedia;
            }

            if (this.conversationsClient) {
                this.outgoingInvitation = this.conversationsClient.inviteToConversation(this.props.userId, options);
                this.outgoingInvitation.then(
                    this.conversationStarted,
                    (error) => {
                        console.error('Unable to create conversation', error); //eslint-disable-line no-console
                        if (this.mounted) {
                            this.setState({
                                error: Utils.localizeMessage(holders.unable.id, holders.unable.defaultMessage)
                            });
                        }
                    }
                );
            }
        } else {
            this.setState({
                isAnswering: true
            });

            this.conversationsClient.on('invite', (invite) => {
                if (invite.from === this.props.userId) {
                    invite.accept().then(this.conversationStarted);
                } else {
                    invite.reject();

                    this.setState({
                        isAnswering: false
                    });

                    Websockets.sendMessage({
                        channel_id: this.state.channelId,
                        action: Constants.SocketEvents.VIDEO_CALL_REJECT,
                        props: {
                            from_id: this.state.callerId,
                            to_id: this.state.currentUserId
                        }
                    });
                }
            });
        }
    }

    conversationStarted(conversation) {
        const self = this;

        if (this.outgoingInvitation) {
            this.outgoingInvitation = null;
        }

        this.conversation = conversation;
        conversation.on('participantConnected', (participant) => {
            console.log('connected from video call'); //eslint-disable-line no-console

            if (this.mounted) {
                self.mainMediaToLocal();
                participant.media.attach('#main-video');

                self.setState({
                    callInProgress: true,
                    isCalling: false,
                    isAnswering: false
                });
                $('#icons').removeClass('hidden').addClass('active');
            }
        });

        conversation.on('participantFailed', (participant) => {
            console.error(Utils.displayUsername(participant.identity) + ' failed to join the Conversation'); //eslint-disable-line no-console
            this.handleFailedCall();
        });

        conversation.on('disconnected', () => {
            console.log('disconnected from video call'); //eslint-disable-line no-console
            if (this.conversation && this.mounted) {
                self.localMediaToMain();
                $('#icons').removeClass('active').addClass('hidden');
                self.unregister();
            }
        });
    }

    resize() {
        $('.post-right__scroll').scrollTop(100000);
    }

    localMediaToMain() {
        if (this.localMedia) {
            this.localMedia.detach('#local-video');
            this.localMedia.attach('#main-video');
        } else {
            $('#local-video').html('');
            this.previewVideo(true);
        }
    }

    mainMediaToLocal() {
        if (this.localMedia) {
            this.localMedia.detach('#main-video');
            this.localMedia.attach('#local-video');
        } else {
            this.conversation.localMedia.attach('#local-video');
        }
    }

    render() {
        const currentId = UserStore.getCurrentId();

        let error;
        if (this.state.error) {
            error = (
                <div className='video-call__error'>
                    <div className='form-group has-error'>
                        <label className='control-label'>{this.state.error}</label>
                    </div>
                </div>
            );
        }

        let searchForm;
        if (currentId != null) {
            searchForm = <SearchBox/>;
        }

        let buttons;
        let connecting;
        if (this.state.isCalling) {
            buttons = (
                <svg
                    id='cancel'
                    xmlns='http://www.w3.org/2000/svg'
                    width='48'
                    height='48'
                    viewBox='-10 -10 68 68'
                    onClick={() => this.handleCancelCall()}
                >
                    <circle
                        cx='24'
                        cy='24'
                        r='34'
                    >
                        <title>
                            <FormattedMessage
                                id='video_call.cancel'
                                defaultMessage='Cancel Call'
                            />
                        </title>
                    </circle>
                    <path
                        transform='scale(0.8), translate(6,10)'
                        d='M24 18c-3.21 0-6.3.5-9.2 1.44v6.21c0 .79-.46 1.47-1.12 1.8-1.95.98-3.74 2.23-5.33 3.7-.36.35-.85.57-1.4.57-.55 0-1.05-.22-1.41-.59L.59 26.18c-.37-.37-.59-.87-.59-1.42 0-.55.22-1.05.59-1.42C6.68 17.55 14.93 14 24 14s17.32 3.55 23.41 9.34c.37.36.59.87.59 1.42 0 .55-.22 1.05-.59 1.41l-4.95 4.95c-.36.36-.86.59-1.41.59-.54 0-1.04-.22-1.4-.57-1.59-1.47-3.38-2.72-5.33-3.7-.66-.33-1.12-1.01-1.12-1.8v-6.21C30.3 18.5 27.21 18 24 18z'
                        fill='white'
                    />
                </svg>
            );
        } else if (!this.state.callInProgress && this.state.localMediaLoaded) {
            buttons = (
                <svg
                    id='call'
                    xmlns='http://www.w3.org/2000/svg'
                    width='48'
                    height='48'
                    viewBox='-10 -10 68 68'
                    onClick={() => this.handlePlaceCall()}
                    disabled={UserStore.getStatus(this.props.userId) === 'offline'}
                >
                    <circle
                        cx='24'
                        cy='24'
                        r='34'
                    >
                        <title>
                            <FormattedMessage
                                id='video_call.call'
                                defaultMessage='Call'
                            />
                        </title>
                    </circle>
                    <path
                        transform='scale(0.8), translate(65,20), rotate(120)'
                        d='M24 18c-3.21 0-6.3.5-9.2 1.44v6.21c0 .79-.46 1.47-1.12 1.8-1.95.98-3.74 2.23-5.33 3.7-.36.35-.85.57-1.4.57-.55 0-1.05-.22-1.41-.59L.59 26.18c-.37-.37-.59-.87-.59-1.42 0-.55.22-1.05.59-1.42C6.68 17.55 14.93 14 24 14s17.32 3.55 23.41 9.34c.37.36.59.87.59 1.42 0 .55-.22 1.05-.59 1.41l-4.95 4.95c-.36.36-.86.59-1.41.59-.54 0-1.04-.22-1.4-.57-1.59-1.47-3.38-2.72-5.33-3.7-.66-.33-1.12-1.01-1.12-1.8v-6.21C30.3 18.5 27.21 18 24 18z'
                        fill='white'
                    />
                </svg>
            );
        }

        if (this.state.isCalling || this.state.isAnswering) {
            connecting = (
                <div className='connecting'>
                    <Connecting position='absolute'/>
                </div>
            );
        }

        if (this.state.callInProgress) {
            const onClass = 'on';
            const offClass = 'off';
            let audioOnClass = offClass;
            let audioOffClass = onClass;
            let videoOnClass = offClass;
            let videoOffClass = onClass;

            let audioTitle = (
                <FormattedMessage
                    id='video_call.mute_audio'
                    defaultMessage='Mute Audio'
                />
            );

            let videoTitle = (
                <FormattedMessage
                    id='video_call.pause_video'
                    defaultMessage='Turn off Video'
                />
            );

            if (this.state.isMuted) {
                audioOnClass = onClass;
                audioOffClass = offClass;
                audioTitle = (
                    <FormattedMessage
                        id='video_call.unmute_audio'
                        defaultMessage='Unmute audio'
                    />
                );
            }

            if (this.state.isPaused) {
                videoOnClass = onClass;
                videoOffClass = offClass;
                videoTitle = (
                    <FormattedMessage
                        id='vide_call.unpause_video'
                        defaultMessage='Turn on Video'
                    />
                );
            }

            buttons = (
                <div
                    id='icons'
                    className='hidden'
                >

                    <svg
                        id='mute-audio'
                        xmlns='http://www.w3.org/2000/svg'
                        width='48'
                        height='48'
                        viewBox='-10 -10 68 68'
                        onClick={() => this.onToggleAudio()}
                    >
                        <circle
                            cx='24'
                            cy='24'
                            r='34'
                        >
                            <title>{audioTitle}</title>
                        </circle>
                        <path
                            className={audioOffClass}
                            transform='scale(0.6), translate(17,18)'
                            d='M38 22h-3.4c0 1.49-.31 2.87-.87 4.1l2.46 2.46C37.33 26.61 38 24.38 38 22zm-8.03.33c0-.11.03-.22.03-.33V10c0-3.32-2.69-6-6-6s-6 2.68-6 6v.37l11.97 11.96zM8.55 6L6 8.55l12.02 12.02v1.44c0 3.31 2.67 6 5.98 6 .45 0 .88-.06 1.3-.15l3.32 3.32c-1.43.66-3 1.03-4.62 1.03-5.52 0-10.6-4.2-10.6-10.2H10c0 6.83 5.44 12.47 12 13.44V42h4v-6.56c1.81-.27 3.53-.9 5.08-1.81L39.45 42 42 39.46 8.55 6z'
                            fill='white'
                        />
                        <path
                            className={audioOnClass}
                            transform='scale(0.6), translate(17,18)'
                            d='M24 28c3.31 0 5.98-2.69 5.98-6L30 10c0-3.32-2.68-6-6-6-3.31 0-6 2.68-6 6v12c0 3.31 2.69 6 6 6zm10.6-6c0 6-5.07 10.2-10.6 10.2-5.52 0-10.6-4.2-10.6-10.2H10c0 6.83 5.44 12.47 12 13.44V42h4v-6.56c6.56-.97 12-6.61 12-13.44h-3.4z'
                            fill='white'
                        />
                    </svg>

                    <svg
                        id='mute-video'
                        xmlns='http://www.w3.org/2000/svg'
                        width='48'
                        height='48'
                        viewBox='-10 -10 68 68'
                        onClick={() => this.onToggleVideo()}
                    >
                        <circle
                            cx='24'
                            cy='24'
                            r='34'
                        >
                            <title>{videoTitle}</title>
                        </circle>
                        <path
                            className={videoOffClass}
                            transform='scale(0.6), translate(17,16)'
                            d='M40 8H15.64l8 8H28v4.36l1.13 1.13L36 16v12.36l7.97 7.97L44 36V12c0-2.21-1.79-4-4-4zM4.55 2L2 4.55l4.01 4.01C4.81 9.24 4 10.52 4 12v24c0 2.21 1.79 4 4 4h29.45l4 4L44 41.46 4.55 2zM12 16h1.45L28 30.55V32H12V16z'
                            fill='white'
                        />
                        <path
                            className={videoOnClass}
                            transform='scale(0.6), translate(17,16)'
                            d='M40 8H8c-2.21 0-4 1.79-4 4v24c0 2.21 1.79 4 4 4h32c2.21 0 4-1.79 4-4V12c0-2.21-1.79-4-4-4zm-4 24l-8-6.4V32H12V16h16v6.4l8-6.4v16z'
                            fill='white'
                        />
                    </svg>

                    <svg
                        id='hangup'
                        xmlns='http://www.w3.org/2000/svg'
                        width='48'
                        height='48'
                        viewBox='-10 -10 68 68'
                        onClick={() => this.onCancelCall()}
                    >
                        <circle
                            cx='24'
                            cy='24'
                            r='34'
                        >
                            <title>
                                <FormattedMessage
                                    id='video_call.hangup'
                                    defaultMessage='Hangup'
                                />
                            </title>
                        </circle>
                        <path
                            transform='scale(0.7), translate(11,10)'
                            d='M24 18c-3.21 0-6.3.5-9.2 1.44v6.21c0 .79-.46 1.47-1.12 1.8-1.95.98-3.74 2.23-5.33 3.7-.36.35-.85.57-1.4.57-.55 0-1.05-.22-1.41-.59L.59 26.18c-.37-.37-.59-.87-.59-1.42 0-.55.22-1.05.59-1.42C6.68 17.55 14.93 14 24 14s17.32 3.55 23.41 9.34c.37.36.59.87.59 1.42 0 .55-.22 1.05-.59 1.41l-4.95 4.95c-.36.36-.86.59-1.41.59-.54 0-1.04-.22-1.4-.57-1.59-1.47-3.38-2.72-5.33-3.7-.66-.33-1.12-1.01-1.12-1.8v-6.21C30.3 18.5 27.21 18 24 18z'
                            fill='white'
                        />
                    </svg>

                </div>
            );
        }

        return (
            <div className='post-right__container'>
                <div className='search-bar__container sidebar--right__search-header'>{searchForm}</div>
                <div className='sidebar-right__body'>
                    <VideoCallHeader
                        userId={this.props.userId}
                        onClose={this.handleClose}
                    />
                    <div className='post-right__scroll'>
                        <div id='videos'>
                            <div
                                id='main-video'
                                autoPlay={true}
                            />
                            <div
                                id='local-video'
                                autoPlay={true}
                            />
                        </div>
                        {error}
                        {connecting}
                        <div className='video-call__buttons'>
                            {buttons}
                        </div>
                    </div>
                </div>
            </div>
        );
    }
}

VideoCall.defaultProps = {
};

VideoCall.propTypes = {
    intl: intlShape.isRequired,
    userId: React.PropTypes.string.isRequired,
    isCaller: React.PropTypes.bool.isRequired
};

export default injectIntl(VideoCall);