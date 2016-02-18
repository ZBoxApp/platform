/**
 * Created by enahum on 2/9/16.
 */

import ModalStore from '../../stores/modal_store.jsx';
import TeamStore from '../../stores/team_store.jsx';
import Constants from '../../utils/constants.jsx';
import {FormattedMessage} from 'mm-intl';

var Modal = ReactBootstrap.Modal;
var ActionTypes = Constants.ActionTypes;

export default class ConfigAddonModal extends React.Component {
    constructor(props) {
        super(props);

        this.handleToggle = this.handleToggle.bind(this);
        this.handleHide = this.handleHide.bind(this);
        this.handleMessage = this.handleMessage.bind(this);

        this.selectedList = null;

        this.state = {
            show: false,
            addon: null,
            error: ''
        };
    }

    componentDidMount() {
        ModalStore.addModalListener(ActionTypes.TOGGLE_ADDON_CONFIG_MODAL, this.handleToggle);
        window.addEventListener('message', this.handleMessage);
    }

    componentWillUnmount() {
        ModalStore.removeModalListener(ActionTypes.TOGGLE_ADDON_CONFIG_MODAL, this.handleToggle);
        window.removeEventListener('message', this.handleMessage);
    }

    handleMessage(e) {
        const message = e.data;

        if (message === 'close') {
            this.handleHide();
        }
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

        const team = TeamStore.getCurrent();
        const url = `${this.state.addon.config_url}?team_id=${team.id}&team_name=${team.name}`;

        return (
            <Modal
                show={this.state.show}
                bsSize='large'
                keyboard={false}
            >
                <Modal.Header closeButton={false}>
                    <Modal.Title>
                        <FormattedMessage
                            id='config_addon.title'
                            defaultMessage='Configure your new {addon} Addon'
                            values={{
                                addon: this.state.addon.name
                            }}
                        />
                    </Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <div className='fluidMedia'>
                        <iframe
                            src={url}
                            frameBorder='0'
                            seamless='seamless'
                            sandbox='allow-forms allow-popups allow-scripts'
                            scrolling='auto'
                        ></iframe>
                    </div>
                </Modal.Body>
            </Modal>
        );
    }
}
