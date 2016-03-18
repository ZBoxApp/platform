/**
 * Created by enahum on 3/16/16.
 */

import $ from 'jquery';
import React from 'react';
import {FormattedMessage} from 'react-intl';

export default class ErrorPage extends React.Component {
    componentDidMount() {
        $('body').attr('class', 'sticky white');
    }
    render() {
        return (
            <div className='error__container'>
                <div className='error__icon'>
                    <i className='fa fa-exclamation-triangle'/>
                </div>
                <h2>
                    <FormattedMessage
                        id='error_page.title'
                        defaultMessage='{siteName} needs your help:'
                        values={{
                            siteName: global.window.mm_config.SiteName
                        }}
                    />
                </h2>
                <p>{this.props.message}</p>
                <a onClick={this.props.goBack}>
                    <FormattedMessage
                        id='error_page.back'
                        defaultMessage='Go back to ZBox Now!'
                    />
                </a>
            </div>
        );
    }
}

ErrorPage.propTypes = {
    message: React.PropTypes.string.isRequired,
    goBack: React.PropTypes.func
};