/**
 * Created by enahum on 3/17/16.
 */

import React from 'react';
import {FormattedMessage} from 'react-intl';
import {browserHistory} from 'react-router';

export default class GuestRevoked extends React.Component {
    goHome() {
        browserHistory.push('/signup_team');
    }
    render() {
        return (
            <div className='error__container'>
                <div className='error__icon'>
                    <i className='fa fa-exclamation-triangle'/>
                </div>
                <h2>
                    <FormattedMessage
                        id='guest_revoked.title'
                        defaultMessage='Invitation Revoked:'
                    />
                </h2>
                <p>
                    <FormattedMessage
                        id='guest_revoked.message'
                        defaultMessage="You're no longer allowed to participate in the conversation"
                    />
                </p>
                <a onClick={this.goHome}>
                    <FormattedMessage
                        id='guest_revoked.back'
                        defaultMessage='Create a team in ZBox Now!'
                    />
                </a>
            </div>
        );
    }
}