import React from 'react';

// based on https://rastating.github.io/creating-a-conditional-react-hook/

type useAPIParams = {
  endpoint: string;
  method: string;
  skip: boolean;
};
export default function useAPI({ endpoint, method, skip }: useAPIParams) {
  const [result, setResult] = React.useState<any>(null);
  const [loading, setLoading] = React.useState(false);
  const [hasError, setHasError] = React.useState(false);
  if (!skip && !loading) {
    setLoading(true);
  }

  React.useEffect(() => {
    if (!skip) {
      const executeRequest = async () => {
        try {
          const res = await fetch(endpoint, {
            method,
          });
          setResult(await res.json());
        } catch (error) {
          setHasError(true);
        }
        setLoading(false);
      };
      executeRequest();
    }
  }, [skip, endpoint, method]);

  return [loading, hasError, result];
}
