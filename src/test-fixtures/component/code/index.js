exports.default = {
  key: "componentKey",
  display: {
    label: "Component label",
    description: "Component description",
    iconPath: "icon.png",
  },
  version: "0.0.1",
  actions: {
    actionKey: {
      key: "actionKey",
      display: {
        label: "Action label",
        description: "Action description",
      },
      inputs: [
        {
          key: "inputKey",
          label: "Input label",
          type: "string",
        },
      ],
      perform: async (context, params) => context.logger.warn("oi"),
    },
  },
};
