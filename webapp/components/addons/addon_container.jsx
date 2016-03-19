/**
 * Created by enahum on 2/8/16.
 */

import React from 'react';
import ErrorStore from 'stores/error_store.jsx';
import AddonStore from 'stores/addon_store.jsx';
import AddonItem from './addon_item.jsx';
import {FormattedMessage} from 'react-intl';

export default class AddonContainer extends React.Component {
    constructor(props) {
        super(props);

        this.onAddonsChange = this.onAddonsChange.bind(this);

        this.state = {serverError: this.props.error, addons: this.props.addons, category: this.props.category};
    }

    componentWillMount() {
        if (this.state.serverError) {
            ErrorStore.storeLastError({message: this.state.serverError});
            ErrorStore.emitChange();
        } else {
            ErrorStore.clearLastError();
            ErrorStore.emitChange();
        }
    }

    componentDidMount() {
        AddonStore.addChangeListener(this.onAddonsChange);
    }

    componentWillUnmount() {
        AddonStore.removeChangeListener(this.onAddonsChange);
    }

    componentWillReceiveProps(nextProps) {
        const state = {serverError: nextProps.error, addons: nextProps.addons, category: nextProps.category};
        this.setState(state);
    }

    onAddonsChange() {
        const state = {
            serverError: '',
            category: this.props.category,
            addons: AddonStore.getAddons(this.props.category)
        };

        this.setState(state);
    }

    render() {
        let addons = this.state.addons.map((a) => {
            return (
                <AddonItem
                    key={a.id}
                    addon={a}
                />
           );
        });

        if (addons.length === 0) {
            addons = (
                <div>
                    <FormattedMessage
                        id='addon.container.none'
                        defaultMessage='This Category does not have any available addons :('
                    />
                </div>
            );
        }

        return (
            <div className='wrapper'>
                <h3>
                    <FormattedMessage
                        id='addon.container.title'
                        defaultMessage='Manage Team Addons'
                    />
                </h3>
                <div className='col-sm-12'>
                    {addons}
                </div>
            </div>
        );
    }
}

AddonContainer.propTypes = {
    addons: React.PropTypes.array.isRequired,
    category: React.PropTypes.string.isRequired,
    error: React.PropTypes.string
};