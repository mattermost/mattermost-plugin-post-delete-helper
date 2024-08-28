import {Action, Store} from 'redux';

import {getPost} from 'mattermost-redux/selectors/entities/posts';
import {GlobalState} from 'mattermost-redux/types/store';

import manifest from '@/manifest';

import {PluginRegistry} from '@/types/mattermost-webapp';

import {triggerDeleteRootPostCommand} from './actions';

export default class Plugin {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars, @typescript-eslint/no-empty-function
    public async initialize(registry: PluginRegistry, store: Store<GlobalState, Action<Record<string, unknown>>>) {
        // @see https://developers.mattermost.com/extend/plugins/webapp/reference/
        registry.registerPostDropdownMenuAction(
            'Remove root post',
            async (postID) => {
                store.dispatch(triggerDeleteRootPostCommand(postID) as any);
            },
            (postID) => {
                const state = store.getState();
                const post = getPost(state, postID);

                // Only show menu if the post is a thread root (has reply posts). Permissions are checked server-side.
                return post.reply_count > 0;
            },
        );
    }
}

declare global {
    interface Window {
        registerPlugin(pluginId: string, plugin: Plugin): void
    }
}

window.registerPlugin(manifest.id, new Plugin());
