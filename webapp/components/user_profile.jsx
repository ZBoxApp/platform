// Copyright (c) 2015 Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

import UserStore from 'stores/user_store.jsx';
import * as GlobalActions from 'action_creators/global_actions.jsx';
import * as Utils from 'utils/utils.jsx';

import {FormattedMessage} from 'react-intl';

import {Popover, OverlayTrigger} from 'react-bootstrap';

var id = 0;

function nextId() {
    id = id + 1;
    return id;
}

import React from 'react';

export default class UserProfile extends React.Component {
    constructor(props) {
        super(props);
        this.uniqueId = nextId();

        this.makeCall = this.makeCall.bind(this);
        this.state = {
            currentUserId: UserStore.getCurrentId()
        };
    }
    makeCall() {
        if (UserStore.getStatus(this.props.user.id) !== 'offline') {
            GlobalActions.makeVideoCall(this.props.user.id);
        }
    }
    render() {
        let name = '...';
        let email = '';
        let profileImg = '';
        if (this.props.user) {
            name = Utils.displayUsername(this.props.user.id);
            email = this.props.user.email;
            profileImg = '/api/v1/users/' + this.props.user.id + '/image?time=' + this.props.user.update_at;
        }

        if (this.props.overwriteName) {
            name = this.props.overwriteName;
        }

        if (this.props.overwriteImage) {
            profileImg = this.props.overwriteImage;
        }

        if (this.props.disablePopover) {
            return <div>{name}</div>;
        }

        let makeCall;
        const userMedia = navigator.getUserMedia || navigator.webkitGetUserMedia || navigator.mozGetUserMedia;
        if (global.window.mm_config.EnableTwilio === 'true' && userMedia && this.props.user.id !== this.state.currentUserId) {
            const isOnline = UserStore.getStatus(this.props.user.id) !== 'offline';
            let circleClass = 'offline';
            let offlineClass = 'off';
            if (isOnline) {
                circleClass = '';
                offlineClass = 'on';
            }

            makeCall = (
                <div
                    className='video-call__user-profile'
                    key='makeCall'
                >
                    <a
                        href='#'
                        onClick={() => this.makeCall()}
                        disabled={!isOnline}
                    >
                        <svg
                            id='video-btn'
                            xmlns='http://www.w3.org/2000/svg'
                        >
                            <circle
                                className={circleClass}
                                cx='16'
                                cy='16'
                                r='18'
                            >
                                <title>
                                    <FormattedMessage
                                        id='channel_header.make_video_call'
                                        defaultMessage='Make Video Call'
                                    />
                                </title>
                            </circle>
                            <path
                                className={offlineClass}
                                transform='scale(0.4), translate(17,16)'
                                d='M40 8H15.64l8 8H28v4.36l1.13 1.13L36 16v12.36l7.97 7.97L44 36V12c0-2.21-1.79-4-4-4zM4.55 2L2 4.55l4.01 4.01C4.81 9.24 4 10.52 4 12v24c0 2.21 1.79 4 4 4h29.45l4 4L44 41.46 4.55 2zM12 16h1.45L28 30.55V32H12V16z'
                                fill='white'
                            />
                            <path
                                className='off'
                                transform='scale(0.4), translate(17,16)'
                                d='M40 8H8c-2.21 0-4 1.79-4 4v24c0 2.21 1.79 4 4 4h32c2.21 0 4-1.79 4-4V12c0-2.21-1.79-4-4-4zm-4 24l-8-6.4V32H12V16h16v6.4l8-6.4v16z'
                                fill='white'
                            />
                        </svg>
                    </a>
                </div>
            );
        }

        var dataContent = [];
        dataContent.push(
            <img
                className='user-popover__image'
                src={profileImg}
                height='128'
                width='128'
                key='user-popover-image'
            />
        );

        dataContent.push(makeCall);

        if (!global.window.mm_config.ShowEmailAddress === 'true') {
            dataContent.push(
                <div
                    className='text-nowrap'
                    key='user-popover-no-email'
                >
                    <FormattedMessage
                        id='user_profile.notShared'
                        defaultMessage='Email not shared'
                    />
                </div>
            );
        } else {
            dataContent.push(
                <div
                    data-toggle='tooltip'
                    title={email}
                    key='user-popover-email'
                >
                    <a
                        href={'mailto:' + email}
                        className='text-nowrap text-lowercase user-popover__email'
                    >
                        {email}
                    </a>
                </div>
            );
        }

        return (
            <OverlayTrigger
                trigger='click'
                placement='right'
                rootClose={true}
                overlay={
                    <Popover
                        title={name}
                        id='user-profile-popover'
                    >
                        {dataContent}
                    </Popover>
                }
            >
                <div
                    className='user-popover'
                    id={'profile_' + this.uniqueId}
                >
                    {name}
                </div>
            </OverlayTrigger>
        );
    }
}

UserProfile.defaultProps = {
    user: {},
    overwriteName: '',
    overwriteImage: '',
    disablePopover: false
};
UserProfile.propTypes = {
    user: React.PropTypes.object,
    overwriteName: React.PropTypes.string,
    overwriteImage: React.PropTypes.string,
    disablePopover: React.PropTypes.bool
};
