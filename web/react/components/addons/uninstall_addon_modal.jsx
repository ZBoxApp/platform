/**
 * Created by enahum on 2/9/16.
 */

import ModalStore from '../../stores/modal_store.jsx';
import AppDispatcher from '../../dispatcher/app_dispatcher.jsx';
import Constants from '../../utils/constants.jsx';
import * as Client from '../../utils/client.jsx';
import {FormattedMessage} from 'mm-intl';

var Modal = ReactBootstrap.Modal;
var ActionTypes = Constants.ActionTypes;

export default class UninstallAddonModal extends React.Component {
    constructor(props) {
        super(props);

        this.handleUninstall = this.handleUninstall.bind(this);
        this.handleToggle = this.handleToggle.bind(this);
        this.handleHide = this.handleHide.bind(this);

        this.selectedList = null;

        this.state = {
            show: false,
            addon: null,
            error: ''
        };
    }

    componentDidMount() {
        ModalStore.addModalListener(ActionTypes.TOGGLE_ADDON_UNINSTALL_MODAL, this.handleToggle);
    }

    componentWillUnmount() {
        ModalStore.removeModalListener(ActionTypes.TOGGLE_ADDON_UNINSTALL_MODAL, this.handleToggle);
    }

    componentDidUpdate(prevProps, prevState) {
        if (this.state.show && !prevState.show) {
            setTimeout(() => {
                $(ReactDOM.findDOMNode(this.refs.uninstallAddonBtn)).focus();
            }, 0);
        }
    }

    handleUninstall() {
        const state = this.state;

        this.handleHide();
        Client.uninstallAddon(
            state.addon.id,
            () => {
                AppDispatcher.handleViewAction({
                    type: ActionTypes.ADDON_UNINSTALLED,
                    addon_id: state.addon.id
                });
                this.handleHide();
            },
            (err) => {
                state.error = err.message;
                this.setState(state);
            }
        );
    }

    handleToggle(value, args) {
        this.setState({
            show: value,
            addon: args.addon,
            error: ''
        });
    }

    handleHide() {
        this.setState({show: false});
    }

    render() {
        if (!this.state.addon) {
            return null;
        }

        var error = null;
        if (this.state.error) {
            error = <div className='form-group has-error'><label className='control-label'>{this.state.error}</label></div>;
        }

        return (
            <Modal
                show={this.state.show}
                onHide={this.handleHide}
            >
                <Modal.Header closeButton={true}>
                    <Modal.Title>
                        <FormattedMessage
                            id='uninstall_addon.confirm'
                            defaultMessage='Confirm the Uninstall of {addon}'
                            values={{
                                addon: this.state.addon.name
                            }}
                        />
                    </Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <FormattedMessage
                        id='uninstall_addon.question'
                        defaultMessage='Are you sure you want to uninstall the addon {addon}?'
                        values={{
                            addon: this.state.addon.name
                        }}
                    />
                    <br />
                    <br />
                    {error}
                </Modal.Body>
                <Modal.Footer>
                    <button
                        type='button'
                        className='btn btn-default'
                        onClick={this.handleHide}
                    >
                        <FormattedMessage
                            id='uninstall_addon.cancel'
                            defaultMessage='Cancel'
                        />
                    </button>
                    <button
                        ref='uninstallAddonBtn'
                        type='button'
                        className='btn btn-danger'
                        onClick={this.handleUninstall}
                    >
                        <FormattedMessage
                            id='uninstall_addon.ok'
                            defaultMessage='Uninstall'
                        />
                    </button>
                </Modal.Footer>
            </Modal>
        );
    }
}
