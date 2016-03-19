/**
 * Created by enahum on 2/8/16.
 */

import React from 'react';
import * as GlobalActions from 'action_creators/global_actions.jsx';
import {FormattedMessage} from 'react-intl';

export default class AddonItem extends React.Component {
    constructor(props) {
        super(props);
        this.handleInstall = this.handleInstall.bind(this);
        this.handleUninstall = this.handleUninstall.bind(this);
    }

    handleInstall(e) {
        e.preventDefault();
        GlobalActions.showUninstallAddonModal(this.props.addon);
    }

    handleUninstall(e) {
        e.preventDefault();
        GlobalActions.showInstallAddonModal(this.props.addon);
    }

    render() {
        const addon = this.props.addon;

        let button;
        if (addon.installed) {
            button = (
                <a
                    href='#'
                    className='btn btn-danger btn-emboss'
                    onClick={this.handleInstall}
                >
                    <FormattedMessage
                        id='addon.item.uninstall'
                        defaultMessage='Uninstall Addon'
                    />
                </a>
            );
        } else {
            button = (
                <a
                    href='#'
                    className='btn btn-primary btn-emboss'
                    onClick={this.handleUninstall}
                >
                    <FormattedMessage
                        id='addon.item.install'
                        defaultMessage='Install Addon'
                    />
                </a>
            );
        }

        return (
            <div className='col-sm-4'>
                <div className='item'>
                    <div className='header'>
                        <img
                            src={addon.icon_url}
                            alt={addon.name}
                        />
                        <div className='name'>{addon.name}</div>
                    </div>
                    <div className='description'>{addon.description}</div>
                    <div className='actions'>
                        <div className='row'>
                            <div className='col-sm-12'>
                                {button}
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        );
    }
}

AddonItem.propTypes = {
    addon: React.PropTypes.object.isRequired
};