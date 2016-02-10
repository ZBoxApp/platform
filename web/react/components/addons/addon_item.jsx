/**
 * Created by enahum on 2/8/16.
 */

import * as EventHelpers from '../../dispatcher/event_helpers.jsx';
import {FormattedMessage} from 'mm-intl';

export default class AddonItem extends React.Component {
    constructor(props) {
        super(props);
    }

    render() {
        const addon = this.props.addon;

        let button;
        if (addon.installed) {
            button = (
                <a
                    href='#'
                    className='btn btn-danger btn-emboss'
                    onClick={() => EventHelpers.showUninstallAddonModal(this.props.addon)}
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
                    onClick={() => EventHelpers.showInstallAddonModal(this.props.addon)}
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