## verify documentation is present. 
## cd is /workspace
HOME_MD="./README.assets/HOME.md"
README_MD="./README.md"
if [ -f $HOME_MD ] && [ -f $README_MD ]; then
    echo "docs check PASSED"
    exit 0
fi
echo "docs check FAILED.  Missing either" $README_MD "or" $HOME_MD
exit 1
